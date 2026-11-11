package app

import (
	"context"
	"encoding/binary"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/godyy/gactor"
	"github.com/godyy/ggs/internal/base/consts"
	"github.com/godyy/ggs/internal/infra/actor"
	"github.com/godyy/ggskit/infra/noderouter"
)

func TestMigrationStartRequestTargets(t *testing.T) {
	tests := []struct {
		name    string
		request migrationStartRequest
		want    []string
		err     bool
	}{
		{name: "legacy single target", request: migrationStartRequest{TargetNodeIDs: []string{"w2"}}, want: []string{"w2"}},
		{name: "multiple targets", request: migrationStartRequest{TargetNodeIDs: []string{"w2", "w4"}}, want: []string{"w2", "w4"}},
		{name: "missing", err: true},
		{name: "empty", request: migrationStartRequest{TargetNodeIDs: []string{"w2", ""}}, err: true},
		{name: "duplicate", request: migrationStartRequest{TargetNodeIDs: []string{"w2", "w2"}}, err: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.request.targets()
			if tt.err {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil || !slices.Equal(tt.want, got) {
				t.Fatalf("targets = %v, error = %v; want %v", got, err, tt.want)
			}
		})
	}
}

func TestMigrationPlannerMatchesAgentSelection(t *testing.T) {
	targets := []string{"game-worker/w2", "game-worker/w4"}
	planner := migrationPlanner(targets)
	reversed := migrationPlanner([]string{targets[1], targets[0]})
	selector := noderouter.NewRendezvousSelector()
	selector.Set(map[string][]string{consts.NodeGameWorker: targets}, true)
	counts := make(map[string]int)
	for id := int64(1); id <= 1000; id++ {
		uid := gactor.ActorUID{Category: 1, ID: id}
		var key [8]byte
		binary.NativeEndian.PutUint64(key[:], uint64(id))
		want := selector.Pick(consts.NodeGameWorker, key[:], 1)[0]
		got, err := planner(uid)
		if err != nil || got != want {
			t.Fatalf("player %d: target = %q, error = %v; want %q", id, got, err, want)
		}
		other, err := reversed(uid)
		if err != nil || got != other {
			t.Fatalf("player %d: reversed target = %q, error = %v; want %q", id, other, err, got)
		}
		counts[got]++
	}
	if counts[targets[0]] == 0 || counts[targets[1]] == 0 {
		t.Fatalf("both targets should receive actors: %v", counts)
	}
}

func TestReserveMigrationTargetsRollsBackOnFailure(t *testing.T) {
	targets := []actor.WorkerReadiness{{NodeID: "w2"}, {NodeID: "w4"}, {NodeID: "w5"}}
	var reserved, released []string
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := reserveMigrationTargets(ctx, targets, "migration-1",
		func(_ context.Context, target actor.WorkerReadiness, id string) error {
			if id != "migration-1" {
				t.Fatalf("unexpected migration ID %q", id)
			}
			if target.NodeID == "w5" {
				cancel()
				return errors.New("state changed")
			}
			reserved = append(reserved, target.NodeID)
			return nil
		},
		func(ctx context.Context, target actor.WorkerReadiness, id string) {
			if ctx.Err() != nil || id != "migration-1" {
				t.Fatalf("rollback context = %v, migration ID = %q", ctx.Err(), id)
			}
			released = append(released, target.NodeID)
		},
	)
	if err == nil || !strings.Contains(err.Error(), "w5") {
		t.Fatalf("expected error for w5, got %v", err)
	}
	// 失败目标也要尝试释放，以覆盖 Redis 已写入但响应超时的情况。
	if !slices.Equal(reserved, []string{"w2", "w4"}) || !slices.Equal(released, []string{"w2", "w4", "w5"}) {
		t.Fatalf("reserved = %v, released = %v", reserved, released)
	}
}

func TestReserveMigrationTargetsSuccess(t *testing.T) {
	targets := []actor.WorkerReadiness{{NodeID: "w2"}, {NodeID: "w4"}}
	var reserved []string
	err := reserveMigrationTargets(context.Background(), targets, "migration-1",
		func(_ context.Context, target actor.WorkerReadiness, _ string) error {
			reserved = append(reserved, target.NodeID)
			return nil
		},
		func(context.Context, actor.WorkerReadiness, string) {
			t.Fatal("successful reservation must not be released")
		},
	)
	if err != nil || !slices.Equal(reserved, []string{"w2", "w4"}) {
		t.Fatalf("reserved = %v, error = %v", reserved, err)
	}
}
