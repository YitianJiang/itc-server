package form

type AccessTokenForm struct {
	AppId         *string    `json:"app_id" validate:"required"`
	AppSecret     *string     `json:"app_secret" validate:"required"`
}

type AccessTokenResp struct {
	Code 		 int32    `json:"code" validate:"required"`
	Expire       int32   `json:"expire" validate:"omitempty"`
	TenantAccessToken     string     `json:"tenant_access_token" validate:"omitempty"`
	Error        string     `json:"error" validate:"omitempty"`
}
