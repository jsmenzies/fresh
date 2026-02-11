package domain

type Branches struct {
	Current  Branch
	Merged   []string
	Squashed []string
}
