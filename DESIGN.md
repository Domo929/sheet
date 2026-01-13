# D&D 5e Character Sheet TUI - Design Document

## Project Overview
A Terminal User Interface (TUI) application written in Golang using the Bubble Tea library for managing D&D 5th Edition character sheets. The application provides an interactive, feature-rich interface for character creation, progression, combat tracking, and inventory management.

## Core Features

### 1. Character Management
- **Character Creation**: Full character creation workflow including:
  - Race selection (with racial traits and bonuses)
  - Class selection (with class features and proficiencies)
  - Background selection
  - Ability score assignment:
    - Manual entry (type in values)
    - Standard array (15, 14, 13, 12, 10, 8)
    - Point buy system (27 points)
  - Starting equipment selection
  - Proficiency selection where applicable
- **Character Loading/Saving**: 
  - Save multiple characters as JSON files in user's home directory (e.g., `~/.dnd-characters/`)
  - Load existing characters from file
  - Character list/selection screen
- **Character Progression**:
  - Experience point tracking
  - Automatic level-up prompts when XP threshold reached
  - Level-up workflow (HP increase, new features, ability score improvements/feats)
  - Subclass selection at appropriate levels

### 2. Core Statistics
- **Ability Scores**: STR, DEX, CON, INT, WIS, CHA
  - Display both score and modifier
  - Track temporary modifications
- **Proficiency Bonus**: Auto-calculated based on level
- **Armor Class (AC)**: 
  - Base calculation from DEX modifier
  - Equipment modifiers (armor, shields)
  - Magic item bonuses
  - Other modifiers (spells, features)
- **Hit Points**:
  - Maximum HP tracking
  - Current HP tracking
  - Temporary HP
  - Hit dice pool and recovery
- **Speed**: Base and modified movement speed
- **Inspiration**: Track whether character currently has inspiration

### 3. Skills & Checks
- All 18 D&D 5e skills with associated ability scores
- Proficiency tracking (not proficient, proficient, expertise)
- Skill check interface showing:
  - Ability modifier
  - Proficiency bonus (if applicable)
  - Total modifier
  - Quick action to "roll" (mark for future implementation)

### 4. Saving Throws
- All six saving throw types (STR, DEX, CON, INT, WIS, CHA)
- Proficiency tracking per saving throw
- Display total modifier for each
- Quick action interface for making saves

### 5. Combat Features
- **Weapon Attacks**:
  - Equipped weapons list
  - Attack bonus calculation (ability mod + proficiency if applicable)
  - Damage dice and damage type
  - Properties (finesse, versatile, reach, etc.)
  - Magic weapon bonuses
- **Death Saves**:
  - Success/failure tracking (0-3 each)
  - Reset mechanism
  - Visual indicator of character state
- **Conditions**:
  - Track active conditions (blinded, charmed, frightened, poisoned, etc.)
  - Add/remove conditions
  - Display condition effects (for reference)

### 6. Spellcasting
- **Spell Slots**:
  - Track total and remaining spell slots per level (1-9)
  - Short rest and long rest recovery
- **Known/Prepared Spells**:
  - Spell list management
  - Spell details (level, school, casting time, range, components, duration, description)
  - Mark spells as prepared (for classes that prepare)
  - Ritual tag tracking
- **Spellcasting Ability**:
  - Display spell save DC
  - Display spell attack bonus

### 7. Inventory Management
- **Items**:
  - Name, quantity, description
  - Item type (weapon, armor, consumable, magic item, general)
  - Consumable tracking (uses/charges)
- **Equipment Slots**:
  - Weapons (main hand, off-hand)
  - Armor (head, chest, etc.)
  - Accessories (rings, amulets, etc.)
  - Equipped items affect stats automatically
- **Currency**:
  - Track CP, SP, EP, GP, PP
  - Conversion helper
- **Magic Items**:
  - Attunement tracking (max 3 attuned items)
  - Special properties (free-text description)
  - Stat modifications (+1 AC, +2 weapon, etc.)

### 8. Character Features
- **Racial Traits**: Display race-specific features
- **Class Features**: Track and display class features gained per level
- **Feats**: List of acquired feats with descriptions
- **Proficiencies**:
  - Armor proficiencies
  - Weapon proficiencies
  - Tool proficiencies
  - Language proficiencies
  - Saving throw proficiencies

### 9. Character Information
- **Basic Info**:
  - Character name
  - Player name
  - Race, class, level
  - Alignment
  - Background
- **Personality**:
  - Personality traits
  - Ideals
  - Bonds
  - Flaws
  - Backstory (longer text field)

### 10. Rest Mechanics
- **Short Rest Button**:
  - Restore hit dice (spend to recover HP)
  - Restore short rest abilities
  - Class-specific short rest recovery (Warlock spell slots, Fighter Second Wind, etc.)
- **Long Rest Button**:
  - Restore all HP to maximum
  - Restore all spell slots
  - Restore hit dice (up to half character level, minimum 1)
  - Reset daily abilities
  - Clear exhaustion levels (1 level removed)

## User Interface Structure

### View Organization
The TUI will have multiple navigable views:

1. **Character Selection Screen**: List of saved characters with load/create/delete options
2. **Main Character Sheet**: Primary view showing key stats in organized sections
3. **Inventory View**: Detailed inventory management and equipment
4. **Spellbook View**: Spell management and casting interface
5. **Character Info View**: Backstory, personality, and detailed features
6. **Level Up View**: Guided level-up workflow
7. **Combat View**: Focused view for attacks, death saves, conditions, HP
8. **Rest Dialog**: Modal for taking short/long rests with resource restoration

### Main Character Sheet Layout
The main sheet will be organized into sections:
- **Header**: Name, race, class, level, XP, Inspiration indicator
- **Core Stats Panel**: Ability scores, AC, HP, Speed, Proficiency Bonus
- **Skills Panel**: Compact list of all skills with modifiers
- **Saving Throws Panel**: All saves with modifiers
- **Quick Actions Panel**: Common actions (attack, spell, rest, etc.)

### Interaction Model
- **Keyboard Navigation**: Arrow keys, tab, vim-style keys (hjkl)
- **Action Keys**: Enter to select, 'e' to edit, 'a' to add, 'd' to delete, 'r' for rest, etc.
- **Modal Forms**: Pop-up forms for editing/adding data
- **Contextual Help**: Display available keybindings at bottom of screen

### Visual Style
- Default Bubble Tea styling (black and white terminal output)
- Clean borders and spacing for readability
- Text-based indicators (*, •, [], etc.) for states and selections

## Technical Implementation

### Data Model
- **Character Struct**: Main character data structure
- **Sub-structs**: AbilityScores, Skills, Inventory, Spells, Features, etc.
- **JSON Serialization**: Direct marshal/unmarshal to/from file

### External Data Files
Race and class data will be loaded from external JSON files for easy expansion and modification:
- **races.json**: Race definitions including traits, ability bonuses, proficiencies, speed
- **classes.json**: Class definitions including hit dice, proficiencies, features by level, spell slots
- **spells.json**: Spell database with all 5e spells
- **backgrounds.json**: Background options with proficiencies and features
- **conditions.json**: Condition definitions with mechanical effects
- Data files located in `~/.dnd-data/` or bundled with application

### Bubble Tea Architecture
- **Models**: One model per view (CharacterSheetModel, InventoryModel, etc.)
- **Messages**: Custom messages for state updates (TakeDamage, GainXP, EquipItem, TakeRest, etc.)
- **Commands**: Async operations (file I/O, calculations)
- **Components**: Reusable UI components (stat block, skill list, item picker, etc.)

### Calculation Engine
- Centralized functions for:
  - AC calculation
  - Attack bonus calculation
  - Skill modifier calculation
  - Spell DC/attack calculation
  - XP thresholds and leveling
  - Equipment bonus aggregation
  - Point buy validation
  - Rest resource restoration

### File Management
- Character files stored in `~/.dnd-characters/`
- Data files stored in `~/.dnd-data/` (or bundled with app)
- Naming convention: `charactername.json`
- Auto-save on significant changes
- Manual save option

## Future Enhancements (TODO)

### High Priority
- **Dice Rolling System**: 
  - Integrated dice roller with advantage/disadvantage
  - 4d6 drop lowest and other rolled methods for ability scores
  - Roll history/log
- **Item Weight & Encumbrance**: Track carrying capacity and encumbrance penalties
- **UI Customization**: 
  - Theme colors and color schemes
  - Layout preferences
  - Visible sections customization

### Medium Priority
- **Multiclassing Support**: Handle multiple classes and their interactions
- **Character Export**: Export to PDF or other formats
- **Character Import**: Import from D&D Beyond or other sources
- **Spell Filtering**: Search and filter spells by school, level, class, etc.
- **Enhanced Data Files**: Expand race/class/spell databases with additional options

### Low Priority
- **Initiative Tracking**: Simple initiative tracker for solo play
- **Campaign Notes**: Session notes and quest tracking
- **Party Management**: Manage multiple characters as a party
- **Homebrew Content**: Support for custom races, classes, spells

## Implementation Plan

### Git Workflow

All development follows this workflow:

1. **Branch Naming Convention**:
   - Feature branches: `feature/<descriptive-name>`
   - Bug fixes: `bugfix/<descriptive-name>`
   - Data files: `data/<descriptive-name>`

2. **Development Process**:
   - Create a new branch from `main` for each distinct piece of work
   - Commit changes with clear, descriptive commit messages
   - Push branch to remote repository
   - Open a Pull Request (PR) against `main`
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

### Phase 1: Foundation & Project Setup

**Branch: `feature/project-init`**
- Initialize Go module (`go mod init`)
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
- Create `races.json` with core 5e races:
  - Human, Elf (High, Wood), Dwarf (Mountain, Hill)
  - Halfling (Lightfoot, Stout), Dragonborn
  - Gnome (Forest, Rock), Half-Elf, Half-Orc, Tiefling
- Include all racial traits, ability bonuses, proficiencies, speeds, languages

**Parallel Work Stream 1B: `data/core-classes`**
- Create `classes.json` with core 5e classes:
  - Barbarian, Bard, Cleric, Druid
  - Fighter, Monk, Paladin, Ranger
  - Rogue, Sorcerer, Warlock, Wizard
- Include hit dice, proficiencies, class features by level, spell slot progression

**Parallel Work Stream 1C: `data/spells`**
- Create `spells.json` with all 5e SRD spells
- Include: name, level, school, casting time, range, components, duration, description, classes

**Parallel Work Stream 1D: `data/backgrounds-and-conditions`**
- Create `backgrounds.json` with standard backgrounds:
  - Acolyte, Charlatan, Criminal, Entertainer
  - Folk Hero, Guild Artisan, Hermit, Noble
  - Outlander, Sage, Sailor, Soldier, Urchin
- Create `conditions.json` with all 5e conditions and their effects

---

### Phase 2: Core Data Models & Engine

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

### Phase 3: Basic UI Framework

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

### Phase 4: Character Creation Workflow

**Branch: `feature/character-creation-basic-info`**
- Create character creation wizard UI
- Step 1: Character name, player name
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

### Phase 5: Main Character Sheet View

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

### Phase 6: Combat & HP Management

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

### Phase 7: Rest Mechanics

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

### Phase 8: Inventory System

**Branch: `feature/inventory-view`**
- Inventory view UI layout
- Item list display
- Add/remove/edit items
- Quantity management
- Item type categorization
- Currency tracking and display

**Branch: `feature/equipment-system`**
- Equipment slots UI
- Equip/unequip items
- Weapon slots (main hand, off-hand)
- Armor slots
- Accessory slots
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

### Phase 9: Spellcasting System

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

### Phase 10: Character Progression

**Branch: `feature/xp-tracking`**
- XP display on main sheet
- Add XP button/modal
- Level-up notification when threshold reached
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

### Phase 11: Character Information View

**Branch: `feature/character-info-view`**
- Character info view UI layout
- Display personality traits, ideals, bonds, flaws
- Backstory text display
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

### Phase 12: Polish & Integration

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

### Summary of Parallel Work Opportunities

- **Phase 1**: Data files can all be created in parallel (4 branches)
- **Phase 2**: Data loader, calculation engine, and storage can be developed in parallel (3 branches)
- **Phase 4**: Starting equipment and proficiency selection can be done in parallel (2 branches)
- **Phase 5**: Skills and saving throws panels can be developed in parallel (2 branches)
- **Phase 6**: Conditions and weapon attacks can be done in parallel (2 branches)
- **Phase 8**: Magic items and currency management can be developed in parallel (2 branches)
- **Phase 9**: Spell preparation and spell stats can be done in parallel (2 branches)
- **Phase 11**: Features display and proficiencies display can be done in parallel (2 branches)

This allows for significant parallel development while maintaining clear separation of concerns and manageable PR sizes.
