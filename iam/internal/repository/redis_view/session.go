package redisview

type SessionRedisView struct {
	Uuid      string `redis:"uuid"`
	UserUuid  string `redis:"user_uuid"`
	Login     string `redis:"login"`
	CreatedAt int64  `redis:"created_at"`
	ExpiresAt int64  `redis:"expires_at"`
}
