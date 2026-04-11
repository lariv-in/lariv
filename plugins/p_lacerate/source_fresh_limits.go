package p_lacerate

// sourceDefaultMaxFreshPosts is the cap on new [Intel] rows per fetch when the stored value is zero (unset / legacy row).
const sourceDefaultMaxFreshPosts = 25

func sourceMaxFreshPostsForSave(v uint) uint {
	if v == 0 {
		return sourceDefaultMaxFreshPosts
	}
	return v
}

func sourceEffectiveMaxFreshPosts(stored uint) int {
	n := stored
	if n == 0 {
		n = sourceDefaultMaxFreshPosts
	}
	return int(n)
}
