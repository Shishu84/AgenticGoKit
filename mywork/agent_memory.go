package agenticgokit

// Memory represents a simple agent memory system.
// It stores a history of string entries that can be remembered, recalled, or cleared.
type Memory struct {
    History []string
}

// Remember adds a new entry to the memory.
func (m *Memory) Remember(entry string) {
    m.History = append(m.History, entry)
}

// Recall returns all stored entries in memory.
func (m *Memory) Recall() []string {
    return m.History
}

// Forget clears all entries from memory.
func (m *Memory) Forget() {
    m.History = []string{}
}
