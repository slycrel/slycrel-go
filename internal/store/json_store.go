package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"github.com/slycrel/slycrel/internal/model"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrUnauthorized = errors.New("authentication failed")
)

// JSONStore implements Store using JSON files on disk.
type JSONStore struct {
	dataDir  string // path to data/ directory (static game data)
	stateDir string // path to state/ directory (mutable game state)
	mu       sync.RWMutex
}

// NewJSONStore creates a new JSON file store.
// dataDir contains static game data (monsters, weapons, etc.).
// stateDir contains mutable game state (characters, inn, etc.).
func NewJSONStore(dataDir, stateDir string) (*JSONStore, error) {
	// Ensure state directories exist
	dirs := []string{
		stateDir,
		filepath.Join(stateDir, "mail"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("creating state directory %s: %w", dir, err)
		}
	}

	return &JSONStore{
		dataDir:  dataDir,
		stateDir: stateDir,
	}, nil
}

// --- Characters ---

func (s *JSONStore) loadCharacters() ([]model.Character, error) {
	path := filepath.Join(s.stateDir, "characters.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Character{}, nil
		}
		return nil, err
	}
	var chars []model.Character
	if err := json.Unmarshal(data, &chars); err != nil {
		return nil, err
	}
	return chars, nil
}

func (s *JSONStore) saveCharacters(chars []model.Character) error {
	data, err := json.MarshalIndent(chars, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(filepath.Join(s.stateDir, "characters.json"), data)
}

func (s *JSONStore) LoadCharacter(bbsName string) (*model.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.findCharByBBSName(bbsName)
}

func (s *JSONStore) SaveCharacter(char *model.Character) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chars, err := s.loadCharacters()
	if err != nil {
		return err
	}

	found := false
	for i, c := range chars {
		if strings.EqualFold(c.BBSName, char.BBSName) {
			chars[i] = *char
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("character %q not found", char.BBSName)
	}
	return s.saveCharacters(chars)
}

func (s *JSONStore) FindCharacterByName(slyName string) (*model.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chars, err := s.loadCharacters()
	if err != nil {
		return nil, err
	}
	for _, c := range chars {
		if strings.EqualFold(c.Name, slyName) {
			return &c, nil
		}
	}
	return nil, ErrNotFound
}

func (s *JSONStore) FindCharacterByBBSName(bbsName string) (*model.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.findCharByBBSName(bbsName)
}

func (s *JSONStore) findCharByBBSName(bbsName string) (*model.Character, error) {
	chars, err := s.loadCharacters()
	if err != nil {
		return nil, err
	}
	for _, c := range chars {
		if strings.EqualFold(c.BBSName, bbsName) {
			return &c, nil
		}
	}
	return nil, ErrNotFound
}

func (s *JSONStore) ListCharacters() ([]model.Character, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadCharacters()
}

func (s *JSONStore) AddCharacter(char *model.Character) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chars, err := s.loadCharacters()
	if err != nil {
		return err
	}

	// Assign file index
	var maxIndex int64
	for _, c := range chars {
		if c.FileIndex > maxIndex {
			maxIndex = c.FileIndex
		}
	}
	char.FileIndex = maxIndex + 1

	chars = append(chars, *char)
	return s.saveCharacters(chars)
}

func (s *JSONStore) DeleteCharacter(bbsName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chars, err := s.loadCharacters()
	if err != nil {
		return err
	}

	for i, c := range chars {
		if strings.EqualFold(c.BBSName, bbsName) {
			chars = append(chars[:i], chars[i+1:]...)
			return s.saveCharacters(chars)
		}
	}
	return ErrNotFound
}

// --- Monsters ---

func (s *JSONStore) LoadMonsters(region string) ([]model.Monster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dataDir, "monsters", region+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Monster{}, nil
		}
		return nil, err
	}
	var monsters []model.Monster
	return monsters, json.Unmarshal(data, &monsters)
}

func (s *JSONStore) GetRandomMonster(region string, playerLevel int) (*model.Monster, error) {
	monsters, err := s.LoadMonsters(region)
	if err != nil {
		return nil, err
	}
	if len(monsters) == 0 {
		return nil, ErrNotFound
	}

	// Filter to monsters near the player's level
	var candidates []model.Monster
	for _, m := range monsters {
		if m.Level <= playerLevel+2 && m.Level >= playerLevel-2 {
			candidates = append(candidates, m)
		}
	}
	if len(candidates) == 0 {
		candidates = monsters
	}

	pick := candidates[rand.Intn(len(candidates))]
	return &pick, nil
}

// --- Equipment ---

func (s *JSONStore) LoadWeapons() ([]model.Weapon, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return loadJSON[[]model.Weapon](filepath.Join(s.dataDir, "weapons.json"))
}

func (s *JSONStore) LoadArmor() ([]model.Armor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return loadJSON[[]model.Armor](filepath.Join(s.dataDir, "armor.json"))
}

// --- Terrain ---

func (s *JSONStore) LoadTerrain(name string) (*model.TerrainMap, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.dataDir, "terrain", name+".json")
	return loadJSON[*model.TerrainMap](path)
}

func (s *JSONStore) ListTerrainMaps() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(filepath.Join(s.dataDir, "terrain"))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return names, nil
}

// --- Inn ---

func (s *JSONStore) LoadInn() (*model.InnRec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inn, err := loadJSON[*model.InnRec](filepath.Join(s.stateDir, "inn.json"))
	if err != nil {
		if os.IsNotExist(err) {
			defaultInn := &model.InnRec{Open: true, CurRate: 5}
			return defaultInn, nil
		}
		return nil, err
	}
	return inn, nil
}

func (s *JSONStore) SaveInn(inn *model.InnRec) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return saveJSON(filepath.Join(s.stateDir, "inn.json"), inn)
}

// --- Arena / Gladiator ---

func (s *JSONStore) LoadGladiatorFights() ([]model.GCombatPrefs, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fights, err := loadJSON[[]model.GCombatPrefs](filepath.Join(s.stateDir, "gladiator_fights.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return []model.GCombatPrefs{}, nil
		}
		return nil, err
	}
	return fights, nil
}

func (s *JSONStore) SaveGladiatorFight(fight *model.GCombatPrefs) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fights, _ := loadJSON[[]model.GCombatPrefs](filepath.Join(s.stateDir, "gladiator_fights.json"))
	fights = append(fights, *fight)
	return saveJSON(filepath.Join(s.stateDir, "gladiator_fights.json"), fights)
}

func (s *JSONStore) ClearGladiatorFights() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return saveJSON(filepath.Join(s.stateDir, "gladiator_fights.json"), []model.GCombatPrefs{})
}

func (s *JSONStore) LoadBets() ([]model.GCombatBet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bets, err := loadJSON[[]model.GCombatBet](filepath.Join(s.stateDir, "gladiator_bets.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return []model.GCombatBet{}, nil
		}
		return nil, err
	}
	return bets, nil
}

func (s *JSONStore) SaveBet(bet *model.GCombatBet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	bets, _ := loadJSON[[]model.GCombatBet](filepath.Join(s.stateDir, "gladiator_bets.json"))
	bets = append(bets, *bet)
	return saveJSON(filepath.Join(s.stateDir, "gladiator_bets.json"), bets)
}

func (s *JSONStore) ClearBets() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return saveJSON(filepath.Join(s.stateDir, "gladiator_bets.json"), []model.GCombatBet{})
}

// --- High Scores ---

func (s *JSONStore) LoadHighScores() ([]model.HighScore, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scores, err := loadJSON[[]model.HighScore](filepath.Join(s.stateDir, "highscores.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return []model.HighScore{}, nil
		}
		return nil, err
	}
	return scores, nil
}

func (s *JSONStore) SaveHighScores(scores []model.HighScore) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return saveJSON(filepath.Join(s.stateDir, "highscores.json"), scores)
}

// --- Town Config ---

func (s *JSONStore) LoadTownConfig() (*model.TownConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, err := loadJSON[*model.TownConfig](filepath.Join(s.stateDir, "town.json"))
	if err != nil {
		if os.IsNotExist(err) {
			def := model.DefaultTownConfig()
			return &def, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (s *JSONStore) SaveTownConfig(cfg *model.TownConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return saveJSON(filepath.Join(s.stateDir, "town.json"), cfg)
}

// --- Mail / News ---

func (s *JSONStore) WriteNews(playerBBSName string, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.stateDir, "mail", playerBBSName+"--News.txt")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(message + "\n")
	return err
}

func (s *JSONStore) ReadNews(playerBBSName string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.stateDir, "mail", playerBBSName+"--News.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (s *JSONStore) ClearNews(playerBBSName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.stateDir, "mail", playerBBSName+"--News.txt")
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// --- Helpers ---

func loadJSON[T any](path string) (T, error) {
	var zero T
	data, err := os.ReadFile(path)
	if err != nil {
		return zero, err
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return zero, err
	}
	return result, nil
}

func saveJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return atomicWrite(path, data)
}

// --- Auth ---

// AuthenticatePlayer checks the bcrypt password hash for the given BBS username.
// Returns ErrUnauthorized on any failure (user not found, no password set, wrong password)
// to prevent username enumeration.
func (s *JSONStore) AuthenticatePlayer(bbsName, password string) error {
	s.mu.RLock()
	char, err := s.findCharByBBSName(bbsName)
	s.mu.RUnlock()
	if err != nil || char.PasswordHash == "" {
		return ErrUnauthorized
	}
	if err := bcrypt.CompareHashAndPassword([]byte(char.PasswordHash), []byte(password)); err != nil {
		return ErrUnauthorized
	}
	return nil
}

// SetPlayerPassword hashes and stores a new password for the given BBS username.
func (s *JSONStore) SetPlayerPassword(bbsName, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	chars, err := s.loadCharacters()
	if err != nil {
		return err
	}
	for i, c := range chars {
		if strings.EqualFold(c.BBSName, bbsName) {
			chars[i].PasswordHash = string(hash)
			return s.saveCharacters(chars)
		}
	}
	return ErrNotFound
}

// atomicWrite writes data to a temp file then renames it to path.
func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
