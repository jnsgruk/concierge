package runnertest

func NewTestSnap(name, channel string, classic bool, installed bool) *TestSnap {
	return &TestSnap{name: name, channel: channel, classic: classic, installed: installed}
}

type TestSnap struct {
	name      string
	channel   string
	classic   bool
	installed bool
}

func (ts *TestSnap) Name() string              { return ts.name }
func (ts *TestSnap) Classic() (bool, error)    { return ts.classic, nil }
func (ts *TestSnap) Installed() bool           { return ts.installed }
func (ts *TestSnap) Tracking() (string, error) { return ts.channel, nil }
func (ts *TestSnap) Channel() string           { return ts.channel }
func (ts *TestSnap) SetChannel(c string)       { ts.channel = c }
