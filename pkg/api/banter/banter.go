// Package banter provides idle agent conversation generation for full personality mode
package banter

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/json"
	"fmt"
	mathrand "math/rand"
	"strings"
	"sync"
	"time"
)

//go:embed personalities.json
var personalitiesFS embed.FS

// Agent represents an agent's personality profile for banter
type Agent struct {
	Name                 string              `json:"name"`
	FullName             string              `json:"full_name"`
	Character            string              `json:"character"`
	Domain               string              `json:"domain"`
	Personality          string              `json:"personality"`
	Catchphrases         []string            `json:"catchphrases"`
	MovieQuotes          []string            `json:"movie_quotes"`
	PunTopics            []string            `json:"pun_topics"`
	Puns                 []string            `json:"puns"`
	ConversationStarters []string            `json:"conversation_starters"`
	Reactions            map[string][]string `json:"reactions"`
}

// Rivalry represents a friendly rivalry between agents
type Rivalry struct {
	Agents []string `json:"agents"`
	Topic  string   `json:"topic"`
	Tone   string   `json:"tone"`
}

// Alliance represents agents that work well together
type Alliance struct {
	Agents []string `json:"agents"`
	Topic  string   `json:"topic"`
}

// GroupDynamics defines how agents interact
type GroupDynamics struct {
	Rivalries  []Rivalry  `json:"rivalries"`
	Alliances  []Alliance `json:"alliances"`
	Mentorship []struct {
		Mentor string `json:"mentor"`
		Mentee string `json:"mentee"`
		Topic  string `json:"topic"`
	} `json:"mentorship"`
}

// SharedQuotes contains quotes all agents might use
type SharedQuotes struct {
	HackersMovie []string `json:"hackers_movie"`
	Greetings    []string `json:"greetings"`
	SignOffs     []string `json:"sign_offs"`
}

// Personalities holds all agent personality data
type Personalities struct {
	Agents        map[string]Agent `json:"agents"`
	GroupDynamics GroupDynamics    `json:"group_dynamics"`
	SharedQuotes  SharedQuotes     `json:"shared_quotes"`
}

// BanterMessage represents a single banter message
type BanterMessage struct {
	ID        string    `json:"id"`
	Agent     string    `json:"agent"`
	AgentName string    `json:"agent_name"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // "pun", "quote", "reaction", "conversation"
	Target    string    `json:"target,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Context provides context for banter generation
type Context struct {
	Repo         string            `json:"repo,omitempty"`
	FindingCount int               `json:"finding_count,omitempty"`
	LastAgent    string            `json:"last_agent,omitempty"`
	RecentTopics []string          `json:"recent_topics,omitempty"`
	ScanResults  map[string]int    `json:"scan_results,omitempty"` // scanner -> finding count
	Custom       map[string]string `json:"custom,omitempty"`
}

// Generator generates banter messages
type Generator struct {
	personalities *Personalities
	enabled       bool
	mu            sync.RWMutex
	rng           *mathrand.Rand
}

// NewGenerator creates a new banter generator
func NewGenerator() (*Generator, error) {
	data, err := personalitiesFS.ReadFile("personalities.json")
	if err != nil {
		return nil, err
	}

	var p Personalities
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	return &Generator{
		personalities: &p,
		enabled:       false, // Disabled by default
		rng:           mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
	}, nil
}

// SetEnabled enables or disables banter generation
func (g *Generator) SetEnabled(enabled bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.enabled = enabled
}

// IsEnabled returns whether banter is enabled
func (g *Generator) IsEnabled() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.enabled
}

// GetAgent returns an agent's personality
func (g *Generator) GetAgent(name string) (*Agent, bool) {
	agent, ok := g.personalities.Agents[strings.ToLower(name)]
	return &agent, ok
}

// ListAgents returns all available agent names
func (g *Generator) ListAgents() []string {
	var names []string
	for name := range g.personalities.Agents {
		names = append(names, name)
	}
	return names
}

// GeneratePun generates a random pun from an agent
func (g *Generator) GeneratePun(agentName string) *BanterMessage {
	agent, ok := g.GetAgent(agentName)
	if !ok || len(agent.Puns) == 0 {
		return nil
	}

	return &BanterMessage{
		ID:        generateID(),
		Agent:     strings.ToLower(agentName),
		AgentName: agent.Name,
		Message:   g.randomChoice(agent.Puns),
		Type:      "pun",
		Timestamp: time.Now(),
	}
}

// GenerateQuote generates a Hackers movie quote from an agent
func (g *Generator) GenerateQuote(agentName string) *BanterMessage {
	agent, ok := g.GetAgent(agentName)
	if !ok {
		return nil
	}

	// Mix agent-specific and shared movie quotes
	quotes := append([]string{}, agent.MovieQuotes...)
	quotes = append(quotes, g.personalities.SharedQuotes.HackersMovie...)

	if len(quotes) == 0 {
		return nil
	}

	return &BanterMessage{
		ID:        generateID(),
		Agent:     strings.ToLower(agentName),
		AgentName: agent.Name,
		Message:   g.randomChoice(quotes),
		Type:      "quote",
		Timestamp: time.Now(),
	}
}

// GenerateReaction generates a reaction from an agent
func (g *Generator) GenerateReaction(agentName string, reactionType string) *BanterMessage {
	agent, ok := g.GetAgent(agentName)
	if !ok {
		return nil
	}

	reactions, ok := agent.Reactions[reactionType]
	if !ok || len(reactions) == 0 {
		return nil
	}

	return &BanterMessage{
		ID:        generateID(),
		Agent:     strings.ToLower(agentName),
		AgentName: agent.Name,
		Message:   g.randomChoice(reactions),
		Type:      "reaction",
		Timestamp: time.Now(),
	}
}

// GenerateConversationStarter generates a conversation starter from an agent
func (g *Generator) GenerateConversationStarter(agentName string, ctx *Context) *BanterMessage {
	agent, ok := g.GetAgent(agentName)
	if !ok || len(agent.ConversationStarters) == 0 {
		return nil
	}

	message := g.randomChoice(agent.ConversationStarters)
	message = g.substituteContext(message, ctx)

	return &BanterMessage{
		ID:        generateID(),
		Agent:     strings.ToLower(agentName),
		AgentName: agent.Name,
		Message:   message,
		Type:      "conversation",
		Timestamp: time.Now(),
	}
}

// GenerateExchange generates a back-and-forth exchange between agents
func (g *Generator) GenerateExchange(ctx *Context) []*BanterMessage {
	if !g.IsEnabled() {
		return nil
	}

	agents := g.ListAgents()
	if len(agents) < 2 {
		return nil
	}

	// Pick 2-3 random agents
	g.shuffleStrings(agents)
	numAgents := 2 + g.rng.Intn(2) // 2 or 3 agents
	if numAgents > len(agents) {
		numAgents = len(agents)
	}
	participants := agents[:numAgents]

	var messages []*BanterMessage

	// First agent starts with a conversation starter or pun
	firstAgent := participants[0]
	if g.rng.Float32() < 0.5 {
		if msg := g.GenerateConversationStarter(firstAgent, ctx); msg != nil {
			messages = append(messages, msg)
		}
	} else {
		if msg := g.GeneratePun(firstAgent); msg != nil {
			messages = append(messages, msg)
		}
	}

	// Other agents react
	for i := 1; i < len(participants); i++ {
		agent := participants[i]

		// Determine reaction type
		reactionTypes := []string{"agreement", "disagreement", "findings"}
		reactionType := g.randomChoice(reactionTypes)

		if msg := g.GenerateReaction(agent, reactionType); msg != nil {
			msg.Target = firstAgent
			messages = append(messages, msg)
		}
	}

	// Maybe end with a shared quote
	if g.rng.Float32() < 0.2 && len(messages) > 0 {
		lastAgent := participants[len(participants)-1]
		if msg := g.GenerateQuote(lastAgent); msg != nil {
			messages = append(messages, msg)
		}
	}

	return messages
}

// GenerateIdleBanter generates random idle banter suitable for display
func (g *Generator) GenerateIdleBanter(ctx *Context) *BanterMessage {
	if !g.IsEnabled() {
		return nil
	}

	agents := g.ListAgents()
	if len(agents) == 0 {
		return nil
	}

	agent := g.randomChoice(agents)

	// Randomly select banter type
	roll := g.rng.Float32()
	switch {
	case roll < 0.4:
		return g.GeneratePun(agent)
	case roll < 0.7:
		return g.GenerateConversationStarter(agent, ctx)
	case roll < 0.9:
		return g.GenerateQuote(agent)
	default:
		return g.GenerateReaction(agent, "findings")
	}
}

// substituteContext replaces placeholders with context values
func (g *Generator) substituteContext(message string, ctx *Context) string {
	if ctx == nil {
		ctx = &Context{}
	}

	replacer := strings.NewReplacer(
		"{repo}", orDefault(ctx.Repo, "this repo"),
		"{finding_count}", orDefaultInt(ctx.FindingCount, "some"),
		"{agent}", orDefault(ctx.LastAgent, "everyone"),
	)

	return replacer.Replace(message)
}

// randomChoice returns a random element from a slice
func (g *Generator) randomChoice(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[g.rng.Intn(len(items))]
}

// shuffleStrings shuffles a string slice in place
func (g *Generator) shuffleStrings(items []string) {
	g.rng.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
}

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%x", time.Now().UnixNano())[:8]
	}
	for i, v := range b {
		b[i] = chars[int(v)%len(chars)]
	}
	return string(b)
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func orDefaultInt(n int, def string) string {
	if n == 0 {
		return def
	}
	return fmt.Sprintf("%d", n)
}

// Service manages banter generation and broadcasting
type Service struct {
	generator  *Generator
	broadcast  func(msg *BanterMessage) error
	interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	running    bool
	banterCtx  *Context
}

// NewService creates a new banter service
func NewService(broadcast func(msg *BanterMessage) error) (*Service, error) {
	gen, err := NewGenerator()
	if err != nil {
		return nil, err
	}

	return &Service{
		generator: gen,
		broadcast: broadcast,
		interval:  30 * time.Second, // Default 30 seconds between banter
	}, nil
}

// SetInterval sets the banter interval
func (s *Service) SetInterval(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = d
}

// SetContext sets the banter context
func (s *Service) SetContext(ctx *Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.banterCtx = ctx
}

// Start starts the banter service
func (s *Service) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.mu.Unlock()

	go s.run()
}

// Stop stops the banter service
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
	s.running = false
}

// SetEnabled enables or disables the banter generator
func (s *Service) SetEnabled(enabled bool) {
	s.generator.SetEnabled(enabled)
	if enabled && !s.running {
		s.Start()
	} else if !enabled && s.running {
		s.Stop()
	}
}

// IsEnabled returns whether banter is enabled
func (s *Service) IsEnabled() bool {
	return s.generator.IsEnabled()
}

// Generator returns the underlying generator
func (s *Service) Generator() *Generator {
	return s.generator
}

func (s *Service) run() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if !s.generator.IsEnabled() {
				continue
			}

			s.mu.Lock()
			ctx := s.banterCtx
			s.mu.Unlock()

			// Generate and broadcast banter
			msg := s.generator.GenerateIdleBanter(ctx)
			if msg != nil && s.broadcast != nil {
				_ = s.broadcast(msg)
			}
		}
	}
}
