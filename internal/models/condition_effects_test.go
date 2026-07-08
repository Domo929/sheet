package models

import "testing"

func TestConditionRollEffect_PoisonedAttackAndCheck(t *testing.T) {
	conds := []Condition{ConditionPoisoned}

	adv, dis, reason := ConditionRollEffect(conds, RollCatAttack, "")
	if adv || !dis {
		t.Errorf("poisoned attack: want disadvantage, got adv=%v dis=%v", adv, dis)
	}
	if reason != "Poisoned" {
		t.Errorf("reason = %q, want Poisoned", reason)
	}

	adv, dis, _ = ConditionRollEffect(conds, RollCatAbilityCheck, "")
	if adv || !dis {
		t.Errorf("poisoned check: want disadvantage, got adv=%v dis=%v", adv, dis)
	}

	// Poisoned does not affect saving throws.
	adv, dis, _ = ConditionRollEffect(conds, RollCatSavingThrow, AbilityDexterity)
	if adv || dis {
		t.Errorf("poisoned save: want no effect, got adv=%v dis=%v", adv, dis)
	}
}

func TestConditionRollEffect_InvisibleGrantsAdvantageOnAttack(t *testing.T) {
	adv, dis, reason := ConditionRollEffect([]Condition{ConditionInvisible}, RollCatAttack, "")
	if !adv || dis {
		t.Errorf("invisible attack: want advantage, got adv=%v dis=%v", adv, dis)
	}
	if reason != "Invisible" {
		t.Errorf("reason = %q, want Invisible", reason)
	}
}

func TestConditionRollEffect_RestrainedDexSaveOnly(t *testing.T) {
	conds := []Condition{ConditionRestrained}

	// DEX save: disadvantage.
	_, dis, _ := ConditionRollEffect(conds, RollCatSavingThrow, AbilityDexterity)
	if !dis {
		t.Error("restrained DEX save: want disadvantage")
	}

	// STR save: no effect.
	_, dis, _ = ConditionRollEffect(conds, RollCatSavingThrow, AbilityStrength)
	if dis {
		t.Error("restrained STR save: want no effect")
	}

	// Attack: disadvantage.
	_, dis, _ = ConditionRollEffect(conds, RollCatAttack, "")
	if !dis {
		t.Error("restrained attack: want disadvantage")
	}
}

func TestConditionRollEffect_MultipleConditionsDedupeReason(t *testing.T) {
	conds := []Condition{ConditionPoisoned, ConditionFrightened, ConditionPoisoned}
	_, dis, reason := ConditionRollEffect(conds, RollCatAttack, "")
	if !dis {
		t.Error("want disadvantage")
	}
	if reason != "Poisoned, Frightened" {
		t.Errorf("reason = %q, want deduped 'Poisoned, Frightened'", reason)
	}
}

func TestConditionRollEffect_NoConditions(t *testing.T) {
	adv, dis, reason := ConditionRollEffect(nil, RollCatAttack, "")
	if adv || dis || reason != "" {
		t.Errorf("no conditions: want clean, got adv=%v dis=%v reason=%q", adv, dis, reason)
	}
}

func TestConditionRollEffect_ProneAndBlindedAttackOnly(t *testing.T) {
	for _, c := range []Condition{ConditionProne, ConditionBlinded} {
		_, dis, _ := ConditionRollEffect([]Condition{c}, RollCatAttack, "")
		if !dis {
			t.Errorf("%s attack: want disadvantage", c)
		}
		_, dis, _ = ConditionRollEffect([]Condition{c}, RollCatAbilityCheck, "")
		if dis {
			t.Errorf("%s check: want no effect", c)
		}
	}
}
