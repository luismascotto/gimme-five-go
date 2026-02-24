// gimme-five-go: CLI that picks a random 5-letter word for Wordle-like games,
// with a roulette-style reveal. Words are loaded once from embedded words_alpha.txt.
package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

//go:embed words_alpha.txt
var wordsAlphaTxt []byte

// fiveLetterWords is populated once at startup from the embedded file.
var fiveLetterWords []string

// Roll delays (ms): accelerate, sustain, then slow to stop (roulette feel).
var rollDelaysMs = []int{1000, 900, 800, 700, 600, 500, 400, 400, 400, 450, 550, 680, 800, 1000, 1500, 2000}

const wordsPerRound = 16

func init() {
	sc := bufio.NewScanner(bytes.NewReader(wordsAlphaTxt))
	for sc.Scan() {
		w := strings.TrimSpace(sc.Text())
		if len(w) == 5 && isAlpha(w) {
			fiveLetterWords = append(fiveLetterWords, strings.ToLower(w))
		}
	}
}

func isAlpha(s string) bool {
	for _, c := range s {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

// pool of indices into fiveLetterWords; shuffled once, consumed in order per round.
type pool struct {
	indices []int
	cursor  int
}

func newPool() *pool {
	n := len(fiveLetterWords)
	idx := make([]int, n)
	for i := 0; i < n; i++ {
		idx[i] = i
	}
	rand.Shuffle(n, func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	return &pool{indices: idx, cursor: 0}
}

func (p *pool) ensureCapacity(need int) {
	remaining := len(p.indices) - p.cursor
	if remaining >= need {
		return
	}
	// Refill: new shuffle and reset cursor
	n := len(fiveLetterWords)
	idx := make([]int, n)
	for i := 0; i < n; i++ {
		idx[i] = i
	}
	rand.Shuffle(n, func(i, j int) { idx[i], idx[j] = idx[j], idx[i] })
	p.indices = idx
	p.cursor = 0
}

func (p *pool) take(n int) []int {
	p.ensureCapacity(n)
	out := make([]int, n)
	copy(out, p.indices[p.cursor:p.cursor+n])
	p.cursor += n
	return out
}

// --- Model & messages ---

type rollTickMsg struct{ t time.Time }
type startRoundMsg struct{}

type model struct {
	words    []string // all 5-letter words
	pool     *pool    // shuffled indices
	state    string   // "rolling" | "stopped"
	roundIdx []int    // indices for current round (len 16)
	step     int      // 0..15 during roll
}

func initialModel() model {
	return model{
		words:    fiveLetterWords,
		pool:     newPool(),
		state:    "rolling",
		roundIdx: nil,
		step:     -1,
	}
}

func (m model) Init() tea.Cmd {
	// Trigger round start on first frame so we can set roundIdx and schedule first tick.
	return tea.Tick(0, func(time.Time) tea.Msg { return startRoundMsg{} })
}

// beginRound prepares the next 16 indices and returns the first tick Cmd.
func (m *model) beginRound() tea.Cmd {
	m.pool.ensureCapacity(wordsPerRound)
	m.roundIdx = m.pool.take(wordsPerRound)
	m.step = 0
	m.state = "rolling"
	return tea.Tick(time.Duration(rollDelaysMs[0])*time.Millisecond, func(t time.Time) tea.Msg {
		return rollTickMsg{t: t}
	})
}

func (m model) currentWord() string {
	if len(m.roundIdx) == 0 || m.step < 0 {
		return ""
	}
	idx := m.roundIdx[m.step]
	if idx >= len(m.words) {
		return ""
	}
	return m.words[idx]
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case startRoundMsg:
		return m, m.beginRound()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "enter":
			if m.state == "stopped" {
				cmd := m.beginRound()
				return m, cmd
			}
			return m, nil
		default:
			return m, nil
		}

	case tea.MouseMsg:
		btn := msg.Button
		if (btn == tea.MouseButtonWheelUp || btn == tea.MouseButtonWheelDown) && m.state == "stopped" {
			cmd := m.beginRound()
			return m, cmd
		}
		return m, nil

	case rollTickMsg:
		m.step++
		if m.step >= wordsPerRound {
			m.step = wordsPerRound - 1
			m.state = "stopped"
			return m, nil
		}
		delayMs := rollDelaysMs[m.step]
		return m, tea.Tick(time.Duration(delayMs)*time.Millisecond, func(t time.Time) tea.Msg {
			return rollTickMsg{t: t}
		})
	}

	return m, nil
}

var (
	wordStyleRolling = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E8E8E8")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 2).
			Margin(1, 0)
	wordStyleFinal = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF87")).
			Background(lipgloss.Color("#0D1B2A")).
			Padding(0, 2).
			Margin(1, 0)
	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)
)

func (m model) View() string {
	w := m.currentWord()
	if w == "" && m.state == "stopped" && len(m.roundIdx) > 0 {
		w = m.words[m.roundIdx[wordsPerRound-1]]
	}
	if w == "" {
		w = "-----"
	}

	var style lipgloss.Style
	if m.state == "rolling" {
		style = wordStyleRolling
	} else {
		style = wordStyleFinal
	}

	// Fixed-width block so the word stays in the same place during roll
	block := style.Render(strings.ToUpper(w))
	hint := hintStyle.Render("Enter or scroll → new round   ·   q / Esc → quit")
	return lipgloss.Place(80, 12, lipgloss.Center, lipgloss.Center, block+"\n\n"+hint, lipgloss.WithWhitespaceChars(" "))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
