// Package concurrent @Author larry
// @Date 2025/6/18 10:29
// @Desc

package concurrent

import (
	"golang.org/x/sync/errgroup"
)

type CGroup struct {
	group errgroup.Group
}

// NewCGroup creates a new CGroup instance.
func NewCGroup(limit int) *CGroup {
	cgroup := &CGroup{}
	cgroup.group.SetLimit(limit)
	return cgroup
}

func (g *CGroup) Go(f func() error) {
	g.group.Go(f)
}

func (g *CGroup) Wait() error {
	return g.group.Wait()
}
