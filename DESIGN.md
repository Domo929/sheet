# D&D 5e Character Sheet TUI - Design Document

## Project Overview
A Terminal User Interface (TUI) application written in Golang using the Bubble Tea library for managing D&D 5th Edition character sheets. The application provides an interactive, feature-rich interface for character creation, progression, combat tracking, and inventory management.

## Core Features

### 1. Character Management
- **Character Creation**: Full character creation workflow including:
  - Progression type selection (XP tracking or Milestone leveling)
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
  - Save multiple characters as JSON files in user's home directory (e.g., `~/.dnd_sheet/characters/`)
  - Load existing characters from file
  - Character list/selection screen
- **Character Progression**:
  - Experience point tracking (if XP-based progression selected)
  - Automatic level-up prompts when XP threshold reached
  - Ability to defer level-up and return to it later
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
  - Ability to recharge items with charges
- **Equipment Slots**:
  - Weapons (main hand, off-hand)
  - Armor slots: head, body armor, cloak/cape, gloves/vambraces, boots
  - Accessories: amulet/necklace (singular), rings (multiple)
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
  - Notes section (general notes, session notes, quest tracking)

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
- **races.json**: Race definitions including traits, ability bonuses, proficiencies, speed (2024 5e data)
- **classes.json**: Class definitions including hit dice, proficiencies, features by level, spell slots (2024 5e data)
- **spells.json**: Spell database with all 5e spells (2024 5e data)
- **backgrounds.json**: Background options with proficiencies and features
- **conditions.json**: Condition definitions with mechanical effects
- Data files located in `~/.dnd_sheet/data/` or bundled with application

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
- Character files stored in `~/.dnd_sheet/characters/`
- Data files stored in `~/.dnd_sheet/data/` (or bundled with app)
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
- **Spell Filtering**: Search and filter spells by school, level, class, etc.
- **UI Customization**: 
  - Theme colors and color schemes
  - Layout preferences
  - Visible sections customization

### Medium Priority
- **Remote Data Source**: 
  - Host race/class/spell/item data files remotely for space efficiency
  - Single source of truth for multiple app instances
  - Optional local caching with compression
  - Fallback to bundled data if remote unavailable
- **Multiclassing Support**: Handle multiple classes and their interactions
- **Character Export**: Export to PDF or other formats
- **Character Import**: Import from D&D Beyond or other sources
- **Enhanced Data Files**: Expand race/class/spell databases with additional options

### Low Priority
- **Initiative Tracking**: Simple initiative tracker for solo play
- **Campaign Notes**: Session notes and quest tracking
- **Party Management**: Manage multiple characters as a party
- **Homebrew Content**: Support for custom races, classes, spells

## Implementation Plan

See [IMPLEMENTATION.md](IMPLEMENTATION.md) for the detailed implementation plan, including:
- Git workflow requirements
- Phased development plan with 12 phases
- Parallel work stream opportunities
- Branch naming conventions
- PR requirements and commit guidelines

