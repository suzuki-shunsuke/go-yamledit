package mag

// RenameKey returns a MapAction renaming given keys from a map.
func RenameKey(key, newKey any) MapAction {
	return &EditMapValueAction{
		Match: MatchMappingValueByKey(key),
		Edit:  EditMappingValueStatic(newKey, NoChange),
	}
}
