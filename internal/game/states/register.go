package states

import "github.com/slycrel/slycrel/internal/game"

// RegisterAll registers all game states with the state machine.
func RegisterAll(sm *game.StateMachine) {
	// Begin / entry flow
	sm.Register(BeginState{})
	sm.Register(EnterSlycrelState{})
	sm.Register(QuitState{})
	sm.Register(DeadState{})

	// Character creation
	sm.Register(CreateCharacterState{})
	sm.Register(CreateCharMenuState{})
	sm.Register(EnterTownFromCreateState{})
	sm.Register(GetCharNameState{})
	sm.Register(OccupationMenuState{})
	sm.Register(FavoriteColorState{})
	sm.Register(ViewCharacterState{})

	// Town
	sm.Register(TownState{})
	sm.Register(TownPromptState{})

	// Wilderness
	sm.Register(WildernessState{})
	sm.Register(WildernessPromptState{})
	sm.Register(SetupCombatState{})

	// Text Combat
	sm.Register(SetupTextCombatState{})
	sm.Register(TextCombatLoopState{})
	sm.Register(UserAttackStageState{})
	sm.Register(OpponentAttackStageState{})
	sm.Register(UserKilledState{})
	sm.Register(UserVictoriousState{})
	sm.Register(UserEscapesState{})
	sm.Register(MonsterRunsState{})

	// Bank
	sm.Register(BankState{})
	sm.Register(BankPromptState{})
	sm.Register(BankDepositState{})
	sm.Register(BankWithdrawState{})

	// Healer
	sm.Register(HealerState{})
	sm.Register(HealerPromptState{})
	sm.Register(AgathaHealsState{})
	sm.Register(HealAmountState{})

	// Herbalist
	sm.Register(HerbalistState{})
	sm.Register(HerbalistPromptState{})
	sm.Register(HerbalistHealState{})

	// Armory
	sm.Register(ArmoryState{})
	sm.Register(ArmoryPromptState{})
	sm.Register(BuyWeaponState{})
	sm.Register(BuyArmorState{})
	sm.Register(SellWeaponState{})
	sm.Register(SellArmorState{})

	// Common Guild
	sm.Register(CommonGuildState{})
	sm.Register(GuildPromptState{})
	sm.Register(TryNewLevelState{})

	// Tavern
	sm.Register(TavernState{})
	sm.Register(TavernPromptState{})
	sm.Register(OrderDrinkState{})
	sm.Register(TalkToLynxState{})
	sm.Register(HangAroundState{})
	sm.Register(CorennePromptState{})
	sm.Register(ViewGuildsState{})

	// Inn
	sm.Register(InnState{})
	sm.Register(InnPromptState{})
	sm.Register(RentRoomState{})
	sm.Register(InnOwnerMenuState{})

	// Arena
	sm.Register(ArenaState{})
	sm.Register(ArenaPromptState{})
	sm.Register(ArenaChallengeState{})
	sm.Register(ArenaGladiatorState{})
	sm.Register(ArenaBetState{})
	sm.Register(ArenaListPlayersState{})
	sm.Register(ArenaChallengeWonState{})
	sm.Register(ArenaChallengeLostState{})

	// Grid Combat
	sm.Register(GridCombatSetupState{})
	sm.Register(GridCombatPromptState{})
	sm.Register(GridMonsterMoveState{})

	// Tower, Blacksmith, Jail
	sm.Register(TowerState{})
	sm.Register(BlacksmithState{})
	sm.Register(JailState{})
}
