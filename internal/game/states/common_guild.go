package states

import (
	"fmt"

	"github.com/slycrel/slycrel/internal/game"
	"github.com/slycrel/slycrel/internal/mechanics"
	"github.com/slycrel/slycrel/internal/model"
)

// CommonGuildState shows the guild ANSI menu.
// Original: ST_CommonGuild
type CommonGuildState struct{}

func (CommonGuildState) ID() game.StateID { return "common_guild" }

func (CommonGuildState) Enter(s *game.Session) {
	s.Character.Location = model.TheCommonGuild
	s.IO.ShowANSIFile("common_guild_menu")
	s.SetNextState("guild_prompt")
}

// GuildPromptState shows the guild menu.
// Original: ST_GuildShortPrompt + ST_GuildMenu
type GuildPromptState struct{}

func (GuildPromptState) ID() game.StateID { return "guild_prompt" }

func (GuildPromptState) Enter(s *game.Session) {
	s.IO.Outln("[L]evel Up, [Q]uit, [?]Help", true, 1)
	s.IO.Cr()

	choice := s.IO.LettersPrompt("Your choice?", "LQ?", 1, true, true)

	switch choice {
	case "L":
		s.SetNextState("try_new_level")
	case "Q":
		s.Character.Location = model.TheTown
		s.SetNextState("town")
	case "?":
		s.SetNextState("common_guild")
	default:
		s.SetNextState("guild_prompt")
	}
}

// TryNewLevelState attempts to advance the character's level.
// Original: ST_TryNewLevel
type TryNewLevelState struct{}

func (TryNewLevelState) ID() game.StateID { return "try_new_level" }

func (TryNewLevelState) Enter(s *game.Session) {
	c := s.Character
	diff := s.TownConfig.LevelUpDifficulty
	if diff < 1 {
		diff = 4
	}

	// Current level and next level requirement
	currentLevel := c.Level()
	required := mechanics.NextLevelUp(currentLevel+1, diff)

	s.IO.Cr()
	s.IO.Outln(fmt.Sprintf("  Your Current Level: %d", currentLevel), true, 4)
	s.IO.Outln(fmt.Sprintf("  Your Total Experience: %d", c.TotalExperience), true, 4)
	s.IO.Outln(fmt.Sprintf("  Required Experience: %d", required), true, 4)
	s.IO.Cr()

	if currentLevel >= 60 {
		s.IO.Outln("The GuildMaster uncomfortably says, \"I'm sorry, you've maxed out.", true, 2)
	} else if c.TotalExperience >= required {
		s.IO.Outln("\"Finally someone who is ready to go up,\" grins the GuildMaster.", true, 5)
		mechanics.GiveNewLevel(c)
		s.IO.Outln(fmt.Sprintf("You have gained one level! You are now level %d!", c.Level()), true, 3)
		s.IO.Cr()

		// Show new stats
		s.IO.Outln(fmt.Sprintf("  HP: %d  Str: %d  Dex: %d  Spd: %d",
			c.MaxHP, c.Strength, c.Dexterity, c.Speed), true, 1)
		if c.CharClass == model.ClassMage {
			s.IO.Outln(fmt.Sprintf("  Psyche: %d", c.MaxPsyche), true, 1)
		}
		s.Store.SaveCharacter(c)
	} else {
		s.IO.Outln("The GuildMaster looks at you and Laughs.", true, 6)
		needed := required - c.TotalExperience
		s.IO.Outln(fmt.Sprintf("  (You need %d more experience)", needed), true, 1)
	}

	s.IO.Cr()
	s.SetNextState("guild_prompt")
}
