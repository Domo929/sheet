# Spellbook Casting

## Overview

The spellbook provides a comprehensive spell management and casting interface for spellcasting characters.

## Casting Spells

1. Open spellbook with 's' key
2. Select a spell from the list (arrow keys or j/k)
3. Press 'c' to cast
4. Review the casting confirmation modal showing:
   - Spell name, level, and school
   - Casting time, range, components, duration
   - Damage (if applicable)
   - Saving throw and DC (if applicable)
   - Full description
   - Upcast effects (for leveled spells)
5. For leveled spells, select slot level with arrow keys
6. Press Enter to cast or Esc to cancel

## Spell Information

The spell details panel (right side) displays:
- Basic spell statistics
- **Damage**: Shows dice and damage type (e.g., "3d6 fire")
- **Saving Throw**: Shows ability and calculated DC (e.g., "Dexterity DC 15")
- Full description
- Upcast information (how the spell improves at higher levels)

## Spell Save DC

The spell save DC is automatically calculated based on:
- Spellcasting ability modifier
- Proficiency bonus
- Formula: 8 + ability modifier + proficiency bonus

## Cantrips

Cantrips can be cast without consuming spell slots. The confirmation modal will show all spell details but no slot selection.

## Pact Magic

Warlocks and other pact magic users will see their pact slots displayed separately from regular spell slots. Pact magic slots are shown with their level (e.g., "Pact Magic - Level 3").
