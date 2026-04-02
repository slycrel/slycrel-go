package store

import "github.com/slycrel/slycrel/internal/model"

// Store abstracts all game data persistence.
// Implementations: JSON file store (now), database (future).
type Store interface {
	// Characters
	LoadCharacter(bbsName string) (*model.Character, error)
	SaveCharacter(char *model.Character) error
	FindCharacterByName(slyName string) (*model.Character, error)
	FindCharacterByBBSName(bbsName string) (*model.Character, error)
	ListCharacters() ([]model.Character, error)
	AddCharacter(char *model.Character) error
	DeleteCharacter(bbsName string) error

	// Monsters
	LoadMonsters(region string) ([]model.Monster, error)
	GetRandomMonster(region string, playerLevel int) (*model.Monster, error)

	// Equipment
	LoadWeapons() ([]model.Weapon, error)
	LoadArmor() ([]model.Armor, error)

	// Terrain
	LoadTerrain(name string) (*model.TerrainMap, error)
	ListTerrainMaps() ([]string, error)

	// Inn
	LoadInn() (*model.InnRec, error)
	SaveInn(inn *model.InnRec) error

	// Arena / Gladiator
	LoadGladiatorFights() ([]model.GCombatPrefs, error)
	SaveGladiatorFight(fight *model.GCombatPrefs) error
	ClearGladiatorFights() error
	LoadBets() ([]model.GCombatBet, error)
	SaveBet(bet *model.GCombatBet) error
	ClearBets() error

	// High Scores
	LoadHighScores() ([]model.HighScore, error)
	SaveHighScores(scores []model.HighScore) error

	// Town config
	LoadTownConfig() (*model.TownConfig, error)
	SaveTownConfig(cfg *model.TownConfig) error

	// Mail / News
	WriteNews(playerBBSName string, message string) error
	ReadNews(playerBBSName string) (string, error)
	ClearNews(playerBBSName string) error
}
