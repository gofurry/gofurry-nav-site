package schema

import "testing"

func TestNormalizeConstraintDefinition(t *testing.T) {
	const legacy = "CHECK (language::text = ANY (ARRAY['zh'::character varying, 'en'::character varying]::text[]))"
	const postgres18 = "CHECK (language::text = ANY (ARRAY['zh'::character varying::text, 'en'::character varying::text]))"

	if actual := normalizeConstraintDefinition("gfn_saying", "chk_gfn_saying_language", legacy); actual != postgres18 {
		t.Fatalf("legacy definition normalized to %q, want %q", actual, postgres18)
	}
	if actual := normalizeConstraintDefinition("another_table", "chk_gfn_saying_language", legacy); actual != legacy {
		t.Fatalf("constraint on another table was changed to %q", actual)
	}
	if actual := normalizeConstraintDefinition("gfn_saying", "another_constraint", legacy); actual != legacy {
		t.Fatalf("another constraint was changed to %q", actual)
	}
}

func TestNormalizeFunctionDefinition(t *testing.T) {
	const identity = "gfg_game_v2_prune_detail_snapshots(p_appid bigint, p_lang text, p_region text, p_keep_count integer)"
	const windowsDefinition = "CREATE FUNCTION probe()\r\nRETURNS integer\r\nAS $function$\r\nBEGIN\r\n  RETURN 1;\r\nEND;\r\n$function$\r\n"
	const canonicalDefinition = "CREATE FUNCTION probe()\nRETURNS integer\nAS $function$\nBEGIN\n  RETURN 1;\nEND;\n$function$\n"

	if actual := normalizeFunctionDefinition(identity, windowsDefinition); actual != canonicalDefinition {
		t.Fatalf("function definition normalized to %q, want %q", actual, canonicalDefinition)
	}
	if actual := normalizeFunctionDefinition("another_function()", windowsDefinition); actual != windowsDefinition {
		t.Fatalf("another function definition was changed to %q", actual)
	}
}
