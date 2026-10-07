package domain

type RefreshToken string
type AccessToken string

func (t AccessToken) String() string  { return string(t) }
func (t RefreshToken) String() string { return string(t) }
