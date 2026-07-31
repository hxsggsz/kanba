package diff

const (
	hScrollStep     = 8
	hScrollFastStep = 32
	VScrollStep     = 3
)

type Scroller struct {
	scroll  int
	hScroll int
}

func NewScroller() *Scroller {
	return &Scroller{}
}

func (s *Scroller) Scroll() int  { return s.scroll }
func (s *Scroller) HScroll() int { return s.hScroll }

// maxScrollFor returns the largest valid scroll offset for content of the
// given total size within a viewport of the given visible size.
func maxScrollFor(total, vis int) int {
	return max(total-vis, 0)
}

func (s *Scroller) MoveDown(total int, vis int) {
	maxScroll := maxScrollFor(total, vis)
	if s.scroll < maxScroll {
		s.scroll = min(s.scroll+VScrollStep, maxScroll)
	}
}

func (s *Scroller) MoveUp() {
	if s.scroll > 0 {
		s.scroll = max(0, s.scroll-VScrollStep)
	}
}

func (s *Scroller) GoToTop() {
	s.scroll = 0
}

func (s *Scroller) GoToBottom(total int, vis int) {
	s.scroll = maxScrollFor(total, vis)
}

func (s *Scroller) ScrollLeft() {
	s.hScroll = max(0, s.hScroll-hScrollStep)
}

func (s *Scroller) ScrollRight() {
	s.hScroll += hScrollStep
}

func (s *Scroller) ScrollLeftFast() {
	s.hScroll = max(0, s.hScroll-hScrollFastStep)
}

func (s *Scroller) ScrollRightFast() {
	s.hScroll += hScrollFastStep
}

func (s *Scroller) ScrollHome() {
	s.hScroll = 0
}

func (s *Scroller) ScrollEnd(maxScroll int) {
	s.hScroll = maxScroll
}

func (s *Scroller) ClampHScroll(maxScroll int) {
	if s.hScroll > maxScroll {
		s.hScroll = max(0, maxScroll)
	}
}

func (s *Scroller) ScrollViewBy(delta int, total int, vis int) {
	if total <= 0 {
		return
	}
	maxScroll := maxScrollFor(total, vis)
	s.scroll = max(0, min(s.scroll+delta, maxScroll))
}

func (s *Scroller) SetScroll(pos, total, vis int) {
	if pos < 0 {
		pos = 0
	}
	maxScroll := maxScrollFor(total, vis)
	if pos > maxScroll {
		pos = maxScroll
	}
	s.scroll = pos
}

func (s *Scroller) UpdateScroll(total int, vis int) {
	if total == 0 || vis <= 0 {
		s.scroll = 0
		return
	}
	maxScroll := maxScrollFor(total, vis)
	if s.scroll > maxScroll {
		s.scroll = maxScroll
	}
	if s.scroll < 0 {
		s.scroll = 0
	}
}
