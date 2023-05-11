package ratelimiter

import "errors"

// ValidateRateLimitConfig Validate the rate limit
// configuration and make sure it is valid and legal
func (c *RateLimitConfig) ValidateRateLimitConfig() error {
	if err := c.validateRedisConfig(); err != nil {
		return err
	}

	if err := c.validateDefaultRateLimitConfig(); err != nil {
		return err
	}

	if err := c.validatePathRateLimitConfig(); err != nil {
		return err
	}

	return nil
}

func (c *RateLimitConfig) validateRedisConfig() error {
	redisConfig := c.RedisConfig
	redisAddress := redisConfig.Address
	if len(redisAddress) == 0 {
		return errors.New("redis address configuration should not be blank or empty")
	}
	return nil
}

func (c *RateLimitConfig) validateDefaultRateLimitConfig() error {
	rateLimitConfigItem := c.DefaultRateLimit
	if rateLimitConfigItem.RateLimitType != LimitTypeBoth {
		return errors.New("rate limit type for the default rate limit config should be LimitByBoth")
	}

	periodRateLimitItem := rateLimitConfigItem.PeriodRateLimitItem
	if periodRateLimitItem.GlobalRateSecond <= 0 || periodRateLimitItem.GlobalRateQuota < 0 {
		return errors.New("globalRateSecond or globalRateQuota in periodRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}
	if periodRateLimitItem.SingleRateSecond <= 0 || periodRateLimitItem.SingleRateQuota < 0 {
		return errors.New("singleRateSecond or singleRateQuota in periodRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}
	if periodRateLimitItem.UserRateSecond <= 0 || periodRateLimitItem.UserRateQuota < 0 {
		return errors.New("userRateSecond or userRateQuota in periodRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}

	tokenRateLimitItem := rateLimitConfigItem.TokenRateLimitItem
	if tokenRateLimitItem.GlobalRate <= 0 || tokenRateLimitItem.GlobalBurst < 0 {
		return errors.New("globalRate or globalBurst in tokenRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}
	if tokenRateLimitItem.SingleRate <= 0 || tokenRateLimitItem.SingleBurst < 0 {
		return errors.New("singleRate or singleBurst in tokenRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}
	return nil
}

func (c *RateLimitConfig) validatePathRateLimitConfig() error {
	pathRateLimitMap := c.PathRateLimitMap
	for _, item := range pathRateLimitMap {
		if item.RateLimitType == LimitTypeBoth {
			if err := validatePeriodRateLimitConfigItem(item.PeriodRateLimitItem); err != nil {
				return err
			}
			if err := validateTokenRateLimitConfigItem(item.TokenRateLimitItem); err != nil {
				return err
			}
		} else if item.RateLimitType == LimitTypePeriod {
			if err := validatePeriodRateLimitConfigItem(item.PeriodRateLimitItem); err != nil {
				return err
			}
		} else if item.RateLimitType == LimitTypeToken {
			if err := validateTokenRateLimitConfigItem(item.TokenRateLimitItem); err != nil {
				return err
			}
		} else {
			return errors.New("rate limit type could only be LimitByPeriod, LimitByToken or LimitByBoth")
		}
	}
	return nil
}

func validatePeriodRateLimitConfigItem(periodRateLimitItem PeriodRateLimitItem) error {
	if periodRateLimitItem.GlobalRateSecond <= 0 || periodRateLimitItem.GlobalRateQuota < 0 {
		return errors.New("globalRateSecond or globalRateQuota in periodRateLimitItem " +
			"for the path rate limit config should be greater than zero")
	}
	if periodRateLimitItem.SingleRateSecond <= 0 || periodRateLimitItem.SingleRateQuota < 0 {
		return errors.New("singleRateSecond or singleRateQuota in periodRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}
	if periodRateLimitItem.UserRateSecond <= 0 || periodRateLimitItem.UserRateQuota < 0 {
		return errors.New("userRateSecond or userRateQuota in periodRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}
	return nil
}

func validateTokenRateLimitConfigItem(tokenRateLimitItem TokenRateLimitItem) error {
	if tokenRateLimitItem.GlobalRate <= 0 || tokenRateLimitItem.GlobalBurst < 0 {
		return errors.New("globalRate or globalBurst in tokenRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}
	if tokenRateLimitItem.SingleRate <= 0 || tokenRateLimitItem.SingleBurst < 0 {
		return errors.New("singleRate or singleBurst in tokenRateLimitItem " +
			"for the default rate limit config should be greater than zero")
	}
	return nil
}
