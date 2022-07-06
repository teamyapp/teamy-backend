package api

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func fromProtoTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}

	tm := ts.AsTime()
	return &tm
}
