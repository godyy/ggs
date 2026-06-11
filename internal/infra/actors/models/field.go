package models

import "go.mongodb.org/mongo-driver/v2/bson"

// ID 通用ID类型约束.
type ID interface {
	int8 | int16 | int32 | int64 | int | uint8 | uint16 | uint32 | uint64 | uint | string
}

// FieldID 通用泛型字段ID.
type FieldID[T ID] struct {
	ID T `bson:"_id"`
}

func (f *FieldID[T]) GetFilter() any {
	return bson.M{"_id": f.ID}
}
