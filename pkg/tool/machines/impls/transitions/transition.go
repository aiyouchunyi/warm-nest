// Package transitions @Author larry
// @Date 2025/5/23 14:30
// @Desc

package transitions

type ITransition interface {
	State() string
}

func AddIfAbsent(transitions []ITransition, transition ITransition) []ITransition {
	for _, t := range transitions {
		if t.State() == transition.State() {
			return transitions
		}
	}
	return append(transitions, transition)
}
