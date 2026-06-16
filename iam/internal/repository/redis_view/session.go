package redisview

type SessionRedisView struct {
	Uuid             string `redis:"uuid"`
	CreatedAt        int64  `redis:"created_at"`
	ExpiresAt        int64  `redis:"expires_at"`
	UserUuid         string `redis:"user_uuid"`
	UserLogin        string `redis:"user_login"`
	UserPasswordHash string `redis:"user_password_hash"`
	UserCreatedAt    int64  `redis:"user_created_at"`
	UserUpdatedAt    int64  `redis:"user_updated_at"`
}
