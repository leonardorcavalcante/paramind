package dedupe

type Set struct {
	items map[string]struct{}
}

func New() *Set {
	return &Set{
		items: make(map[string]struct{}),
	}
}

func (s *Set) Seen(value string) bool {
	if _, exists := s.items[value]; exists {
		return true
	}

	s.items[value] = struct{}{}
	return false
}
