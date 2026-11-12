package domain

type AccessToken string

func (t AccessToken) String() string { return string(t) }
