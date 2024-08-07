package main

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	Key       string
	CreatedAt time.Time
}

type RetentionMap map[string]OTP

func NewRetentionMap(ctx context.Context, retentionPeriod time.Duration) RetentionMap {
	mp := make(RetentionMap)

	go mp.Retention(ctx, retentionPeriod)

	return mp

}

func (m RetentionMap) NewOTP() OTP {
	uid := uuid.NewString()
	o := OTP{Key: uid, CreatedAt: time.Now()}
	m[o.Key] = o
	return o
}

func (m RetentionMap) Verify(otp string) bool {
	_, exists := m[otp]
	if !exists {
		return false
	}
	delete(m, otp)
	return true
}

func (m RetentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) {
	ticker := time.NewTicker(400 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			for _, v := range m {
				if v.CreatedAt.Add(retentionPeriod).Before(time.Now()) {
					delete(m, v.Key)
				}
			}
		case <-ctx.Done():
			return
		}

	}
}
