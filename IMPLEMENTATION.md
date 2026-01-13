# D&D 5e Character Sheet TUI - Implementation Plan

## Git Workflow

All development follows this workflow:

1. **Branch Naming Convention**:
   - Feature branches: `feature/<descriptive-name>`
   - Bug fixes: `bugfix/<descriptive-name>`
   - Data files: `data/<descriptive-name>`

2. **Development Process**:
   - Create a new branch from `main` for each distinct piece of work
   - Commit changes with clear, descriptive commit messages
   - Commits should be appropriately sized and logically grouped for commit-by-commit reviewing
   - Avoid single massive commits with hundreds of files - break them into logical chunks
   - Push branch to remote repository
   - Open a Pull Request (PR) against `main` using `gh pr create`
   - Wait for code review and feedback
   - Iterate on PR based on review comments
   - Once approved, merge PR into `main`

3. **PR Requirements**:
   - Clear description of changes
   - Reference to related tasks/issues
   - Code must compile without errors
   - Basic testing performed (where applicable)

4. **No Direct Commits to Main**:
   - All changes must go through PR process
   - Main branch is protected

---

## Phase 1: Foundation & Project Setup

**Branch: `feature/project-init`**
- Initialize Go module (`go mod init github.com/Domo929/sheet`)
- Add Bubble Tea dependency
- Create basic project structure:
  - `/cmd` - main application entry
  - `/internal` - internal packages
    - `/models` - data models
    - `/ui` - UI components
    - `/engine` - calculation engine
    - `/data` - data loading
    - `/storage` - file management
  - `/data` - external JSON data files
- Create README with build/run instructions
- Add .gitignore for Go projects

**Parallel Work Stream 1A: `data/core-races`**
- Create `races.json` with core 5e races (2024 edition):
  - Human, Elf (High, Wood), Dwarf (Mountain, Hill)
  - Halfling (Lightfoot, Stout), Dragonborn
  - Gnome (Forest, Rock), Half-Elf, Half-Orc, Tiefling
- Include all racial traits, ability bonuses, proficiencies, speeds, languages
- Verify all data matches 2024 5e rules

**Parallel Work Stream 1B: `data/core-classes`**
- Create `classes.json` with core 5e classes (2024 edition):
  - Barbarian, Bard, Cleric, Druid
  - Fighter, Monk, Paladin, Ranger
  - Rogue, Sorcerer, Warlock, Wizard
- Include hit dice, proficiencies, class features by level, spell slot progression
- Verify all data matches 2024 5e rules

**Parallel Work Stream 1C: `data/spells`**
- Create `spells.json` with all 5e SRD spells (2024 edition)
- Include: name, level, school, casting time, range, components, duration, description, classes
- Verify all data matches 2024 5e rules

**Parallel Work Stream 1D: `data/backgrounds-and-conditions`**
- Create `backgrounds.json` with standard backgrounds:
  - Acolyte, Charlatan, Criminal, Entertainer
  - Folk Hero, Guild Artisan, Hermit, Noble
  - Outlander, Sage, Sailor, Soldier, Urchin
- Create `conditions.json` with all 5e conditions and their effects

---

## Phase 2: Core Data Models & Engine

**Branch: `feature/data-models`**
- Define Character struct and all sub-structs:
  - Character, AbilityScores, Skills, SavingThrows
  - Inventory, Item, Equipment, Currency
  - Spellcasting, SpellSlots, Spell
  - Features, Traits, Proficiencies
  - CharacterInfo, Personality
  - CombatStats (HP, AC, DeathSaves, Conditions)
- JSON serialization tags
- Constructor functions and validation

**Parallel Work Stream 2A: `feature/data-loader`**
- Implement data file loading from JSON
- Create loader for races.json, classes.json, spells.json, backgrounds.json, conditions.json
- Error handling and validation
- Data caching in memory

**Parallel Work Stream 2B: `feature/calculation-engine`**
- Implement core calculation functions:
  - Ability modifier calculation
  - Proficiency bonus by level
  - Skill modifier calculation
  - Saving throw modifier calculation
  - AC calculation (base + armor + modifiers)
  - Attack bonus calculation
  - Spell DC and spell attack bonus
  - XP thresholds for leveling
  - Point buy validation and cost calculation

**Parallel Work Stream 2C: `feature/character-storage`**
- Implement character save/load from JSON
- File management (create directory, list characters, delete characters)
- Character file naming and path resolution
- Auto-save functionality
- Error handling for file I/O

---

## Phase 3: Basic UI Framework

**Branch: `feature/ui-framework`**
- Setup base Bubble Tea model structure
- Implement view routing/navigation system
- Create message types for state updates
- Build reusable UI components:
  - Border/panel component
  - List selector component
  - Text input component
  - Button component
  - Help footer component
- Keyboard navigation framework

**Branch: `feature/character-selection-screen`**
- Build character selection/loading screen
- List saved characters
- Create new character option
- Delete character option
- Load character into main app

---

## Phase 4: Character Creation Workflow

**Branch: `feature/character-creation-basic-info`**
- Create character creation wizard UI
- Step 1: Character name, player name, progression type (XP or Milestone)
- Step 2: Race selection with traits display
- Step 3: Class selection with features display
- Step 4: Background selection
- Navigation between steps

**Branch: `feature/ability-score-assignment`**
- Implement ability score assignment step
- Manual entry mode with validation (3-20 range)
- Standard array mode (drag-drop or assign)
- Point buy mode with real-time cost calculation
- Display racial bonuses applied
- Final ability score display with modifiers

**Parallel Work Stream 4A: `feature/starting-equipment`**
- Starting equipment selection UI
- Equipment choices based on class
- Add selected items to inventory
- Set starting gold

**Parallel Work Stream 4B: `feature/proficiency-selection`**
- Skill proficiency selection (based on class/background)
- Tool proficiency selection
- Language selection
- Display proficiency limits and validation

---

## Phase 5: Main Character Sheet View

**Branch: `feature/main-sheet-layout`**
- Build main character sheet view layout
- Header section (name, race, class, level, XP, inspiration)
- Core stats panel (abilities, AC, HP, speed, proficiency)
- Navigation to other views
- Display-only implementation (no editing yet)

**Parallel Work Stream 5A: `feature/skills-panel`**
- Skills panel component
- Display all 18 skills with modifiers
- Show proficiency/expertise indicators
- Highlight calculation (ability + proficiency)

**Parallel Work Stream 5B: `feature/saving-throws-panel`**
- Saving throws panel component
- Display all 6 saves with modifiers
- Show proficiency indicators
- Highlight calculation

---

## Phase 6: Combat & HP Management

**Branch: `feature/hp-tracking`**
- HP display on main sheet
- Damage/healing input modals
- Current HP, Max HP, Temp HP tracking
- Visual HP bar or indicator
- Update character state and save

**Branch: `feature/death-saves`**
- Death save tracking UI
- Success/failure counters
- Mark success/failure buttons
- Reset mechanism
- Visual state indicator (stable, dying, dead)

**Parallel Work Stream 6A: `feature/conditions`**
- Condition tracking component
- Add/remove condition buttons
- Active conditions display
- Condition effect reference viewer

**Parallel Work Stream 6B: `feature/weapon-attacks`**
- Weapon attack panel on combat view
- Display equipped weapons
- Show attack bonus and damage
- Calculate modifiers from stats and proficiency

---

## Phase 7: Rest Mechanics

**Branch: `feature/rest-system`**
- Short rest dialog/modal
- Long rest dialog/modal
- Hit dice spending for short rest
- HP restoration logic
- Spell slot restoration logic
- Hit dice restoration (long rest)
- Resource reset (daily abilities, class features)
- Update character state and save

---

## Phase 8: Inventory System

**Branch: `feature/inventory-view`**
- Inventory view UI layout
- Item list display
- Add/remove/edit items
- Quantity management
- Item type categorization
- Currency tracking and display
- Item charge tracking with recharge functionality

**Branch: `feature/equipment-system`**
- Equipment slots UI
- Equip/unequip items
- Weapon slots (main hand, off-hand)
- Armor slots (head, body armor, cloak/cape, gloves/vambraces, boots)
- Accessory slots (amulet/necklace singular, rings multiple)
- Equipment stat calculation and application to character

**Parallel Work Stream 8A: `feature/magic-items`**
- Attunement tracking (max 3)
- Magic item properties display
- Stat modifiers from magic items
- Special properties (free-text)

**Parallel Work Stream 8B: `feature/currency-management`**
- Currency display (CP, SP, EP, GP, PP)
- Add/remove currency
- Currency conversion helper

---

## Phase 9: Spellcasting System

**Branch: `feature/spellbook-view`**
- Spellbook view UI layout
- Display character's spell list
- Spell details panel
- Add/remove spells

**Branch: `feature/spell-slots`**
- Spell slot tracking display
- Use spell slot button (decrement)
- Restore on rest
- Display by spell level (1-9)

**Parallel Work Stream 9A: `feature/spell-preparation`**
- Mark spells as prepared
- Prepared spell limit calculation (for classes that prepare)
- Known spells vs prepared spells
- Ritual spell indicator

**Parallel Work Stream 9B: `feature/spell-stats`**
- Display spell save DC
- Display spell attack bonus
- Calculate from spellcasting ability
- Update when ability scores or proficiency changes

---

## Phase 10: Character Progression

**Branch: `feature/xp-tracking`**
- XP display on main sheet (if XP-based progression)
- Add XP button/modal
- Level-up notification when threshold reached
- Ability to defer level-up and return later
- XP threshold calculation by level

**Branch: `feature/level-up-workflow`**
- Level-up wizard UI
- HP increase (roll or take average)
- New class features display and selection
- Ability score improvement or feat selection
- New spell slots calculation
- Proficiency bonus update
- Subclass selection at appropriate levels

---

## Phase 11: Character Information View

**Branch: `feature/character-info-view`**
- Character info view UI layout
- Display personality traits, ideals, bonds, flaws
- Backstory text display
- Notes section display
- Edit modals for all fields

**Parallel Work Stream 11A: `feature/features-display`**
- Racial traits display panel
- Class features display panel
- Feats display panel
- Organized by source (race/class/feat)

**Parallel Work Stream 11B: `feature/proficiencies-display`**
- Armor proficiencies list
- Weapon proficiencies list
- Tool proficiencies list
- Language proficiencies list
- Saving throw proficiencies (reference to main sheet)

---

## Phase 12: Polish & Integration

**Branch: `feature/auto-save`**
- Implement auto-save on significant character changes
- Save indicator in UI
- Debouncing to avoid excessive saves

**Branch: `feature/help-documentation`**
- In-app help screen
- Keybinding reference
- Context-sensitive help in each view
- README documentation updates

**Branch: `feature/error-handling`**
- Comprehensive error handling throughout app
- User-friendly error messages
- Validation for all user inputs
- Graceful degradation for missing data files

**Branch: `feature/testing-and-bugfixes`**
- Manual testing of all features
- Bug identification and fixes
- Edge case handling
- Performance testing

---

## Summary of Parallel Work Opportunities

- **Phase 1**: Data files can all be created in parallel (4 branches)
- **Phase 2**: Data loader, calculation engine, and storage can be developed in parallel (3 branches)
- **Phase 4**: Starting equipment and proficiency selection can be done in parallel (2 branches)
- **Phase 5**: Skills and saving throws panels can be developed in parallel (2 branches)
- **Phase 6**: Conditions and weapon attacks can be done in parallel (2 branches)
- **Phase 8**: Magic items and currency management can be developed in parallel (2 branches)
- **Phase 9**: Spell preparation and spell stats can be done in parallel (2 branches)
- **Phase 11**: Features display and proficiencies display can be done in parallel (2 branches)

This allows for significant parallel development while maintaining clear separation of concerns and manageable PR sizes.
