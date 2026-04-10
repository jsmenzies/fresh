package pullrequests

import "fmt"

type Key struct {
	Owner  string
	Repo   string
	Number int
}

func (k Key) String() string {
	return fmt.Sprintf("%s/%s#%d", k.Owner, k.Repo, k.Number)
}

func (k Key) IsValid() bool {
	return k.Owner != "" && k.Repo != "" && k.Number > 0
}
