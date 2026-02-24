# Gimme Five

A small Go CLI that gives you a **random 5-letter word** to start a round of Wordle-like games—with a **roulette-style** reveal so the chosen word feels a bit like a spin of fate.

---

## What it does

On start, the app runs a “roll”: 16 words flash in sequence in the same spot, with delays that speed up, hold, then slow down until the final word stays on screen. That word is your **seed for the round** (e.g. for Wordle, Quordle, Octordle, or any N-Wordle variant). You can roll again with **Enter** or **mouse wheel**, or quit with **q** / **Esc**.

---

## How it works

1. **Startup**
   The word list is read **once** from an **embedded** copy of `words_alpha.txt` (no external file at runtime). Only lines that are exactly 5 letters and alphabetic are kept (~15.9k words).

2. **Pool of indices**
   A **single shuffled slice** of indices `0..N-1` (Fisher–Yates) is built. Each round **consumes the next 16 indices in order**; no “pick random and remove” or “pick until unique,” so there’s no repeated work or bias. When fewer than 16 indices remain, the pool is **re-shuffled** and the cursor reset—rounds can continue indefinitely.

3. **Round**
   For each round, 16 indices are taken from the pool. The app shows the corresponding words **one after another in the same place**, with fixed delays (in ms):
   `1000, 900, 800, 700, 600, 500, 400, 400, 400, 450, 550, 680, 800, 1000, 1500, 2000`
   so the animation **accelerates**, then **slows to a stop** (roulette/slot-machine feel).

4. **UI**
   [Bubble Tea](https://github.com/charmbracelet/bubbletea) drives the TUI; [Lip Gloss](https://github.com/charmbracelet/lipgloss) styles the rolling vs final word and the hint line. The word is drawn in a **fixed area** so the text doesn’t jump during the roll.

---

## Build & run

**Requirements:** Go 1.21+

```bash
# Clone or enter the project directory
cd gimme-five-go

# Build
go build -o gimme-five .

# Run
./gimme-five          # Unix-like
gimme-five.exe        # Windows
```

Or run without building:

```bash
go run .
```

**Controls**

| Action              | Key / input      |
|---------------------|------------------|
| New round           | **Enter** or **mouse wheel** (up/down) |
| Quit                | **q** or **Esc** |

---

## Why it exists & a bit of context

**Purpose**
Many daily word games (Wordle, Quordle, Octordle, etc.) use a single “word of the day” or a fixed sequence. If you want to **play extra rounds** or **pick your own seed** without relying on external generators, this app gives you a **fair, in-order draw** from a large word list, with a small moment of suspense via the roll animation.

**Inspiration**
The idea mixes (1) the need for a **neutral, reproducible source** of 5-letter words with (2) the **slot-machine / roulette** feel—speeding up then slowing to a stop—so the result feels a bit like a mini “spin” instead of an instant random pick.

**Curiosities**
- The word list is **embedded in the binary** (`//go:embed words_alpha.txt`), so after build you don’t need the file on disk; the binary is self-contained.
- The **pool** is a deliberate design: one shuffle, then sequential consumption. It avoids the usual pitfalls (repeated rand + “already seen” checks, or mutating slices by removal) and scales well to large lists.
- The delay curve is **hard-coded** to mimic a simple physical deceleration so the final word feels like it “lands” instead of stopping abruptly.

---

## How Wordle-like (N-Wordle) games work

In the **original Wordle**, you have **6 tries** to guess a hidden **5-letter word**. Each guess is scored per letter:

- **Green** — letter correct and in the right position
- **Yellow** — letter is in the word but in another position
- **Black/gray** — letter not in the word

You use that feedback to narrow down the word. **N-Wordle** is the family of variants that keep the same rules but change the **number of words** (and sometimes board layout):

- **Wordle** — 1 word
- **Dordle** — 2 words
- **Quordle** — 4 words
- **Octordle** — 8 words
- **Sedecordle** — 16 words

You get a few shared tries to narrow things down, plus one extra try per target word (e.g. Quordle gives 9 tries for 4 words).

Typically each game uses a **fixed word (or sequence)** per day. **Gimme Five** doesn’t implement the game itself; it only **picks a 5-letter word** (and can do it again and again) to be used as the opening word (without having to think on the same old tired repetitive ones, like 'apple', 'money', 'trees'). You can use that word as the “target” for your own round of Wordle, Quordle, or any clone you play elsewhere—same rules, your own seed.

---

## Architecture & data choices (highlights)

| Aspect | Choice | Reason |
|--------|--------|--------|
| **Word list** | `//go:embed words_alpha.txt` + one-time parse in `init()` | Single binary, no runtime file I/O; words live in memory for the whole process. |
| **Dictionary** | Filter: length 5, alphabetic only | Matches common Wordle-style valid guess/answer sets. |
| **Randomness** | Pre-shuffled **index pool**, consume in order | Guarantees no duplicate words per pool cycle; O(1) per draw; avoids “random until unique” or repeated slice removal. |
| **Pool refill** | When `remaining < 16`, re-shuffle full index set and reset cursor | Enables unlimited rounds without changing the fairness model. |
| **Roll timing** | Fixed 16 delays (ms), same sequence every round | Predictable “physics”: fast → sustain → slow → stop. |
| **UI** | Bubble Tea (Elm-style) + Lip Gloss | Clear separation of Init/Update/View; easy key and mouse handling; styling and layout in one place. |

---

## Word list

The app expects a file **`words_alpha.txt`** in the project root at **build time** (so it can be embedded). One common source is the [dwyl/english-words](https://github.com/dwyl/english-words) repo (e.g. `words_alpha.txt`). The binary does **not** read this file at runtime; it’s compiled in.

---

## License

Same as the project (if none specified, assume MIT or project default).
