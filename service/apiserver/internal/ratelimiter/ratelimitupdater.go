package ratelimiter

import "github.com/apolloconfig/agollo/v4/storage"

type RateLimitUpdater struct {
}

func (u *RateLimitUpdater) OnChange(event *storage.ChangeEvent) {

}

func (u *RateLimitUpdater) OnNewestChange(event *storage.FullChangeEvent) {

}
