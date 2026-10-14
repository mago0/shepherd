package todo

import (
	"cmp"
	"sort"
	"strings"
)

// Rank maps a priority byte to a sort order (H<M<L<none). Exported for the UI's
// group ordering.
func Rank(p byte) int {
	switch p {
	case 'H':
		return 0
	case 'M':
		return 1
	case 'L':
		return 2
	}
	return 3
}

// catKey sorts named categories alphabetically, uncategorized last.
func catKey(c string) string {
	if c == "" {
		return "￿"
	}
	return strings.ToLower(c)
}

// TagKey is an item's grouping tag: its first tag, or "" when untagged. Tags are
// multi-valued but a row belongs to exactly one group, so the first one wins —
// the rest ride along on the row. Exported for the UI's group ordering.
func TagKey(it Item) string {
	if len(it.Tags) == 0 {
		return ""
	}
	return it.Tags[0]
}

// dueKey sorts soonest-first; no due date sorts last.
func dueKey(d string) string {
	if d == "" {
		return "9999-99-99"
	}
	return d
}

// statusRank is an item's place in the configured statuses ("done" last). A
// status missing from the list sorts just before done. With no configured
// statuses every item ranks the same.
func statusRank(it Item, statuses []string) int {
	if len(statuses) == 0 {
		return 0
	}
	if it.Done {
		return len(statuses)
	}
	name := StatusOf(it, statuses)
	for i, s := range statuses {
		if s == name {
			return i
		}
	}
	return len(statuses) - 1
}

// less is the shared item ordering: overdue pinned first, then category, status
// (in configured order) and priority, or priority, status and category when
// byPrio, then soonest due.
func less(a, b Item, byPrio bool, statuses []string) bool {
	if pa, pb := Pinned(a), Pinned(b); pa != pb {
		return pa
	}
	cat := cmp.Compare(catKey(a.Category), catKey(b.Category))
	status := cmp.Compare(statusRank(a, statuses), statusRank(b, statuses))
	prio := cmp.Compare(Rank(a.Prio), Rank(b.Prio))
	order := [3]int{cat, status, prio}
	if byPrio {
		order = [3]int{prio, status, cat}
	}
	for _, c := range order {
		if c != 0 {
			return c < 0
		}
	}
	return cmp.Compare(dueKey(a.Due), dueKey(b.Due)) < 0
}

// Sort orders items: overdue pinned first, then soonest due. The middle keys
// are category, status, priority, or priority, status, category when byPrio
// (the priority view). statuses is the configured order, "done" last. Each
// item's subtasks are ordered too.
func Sort(items []Item, byPrio bool, statuses []string) {
	sort.SliceStable(items, func(i, j int) bool { return less(items[i], items[j], byPrio, statuses) })
	sortSubs(items, statuses)
}

// sortSubs orders each item's subtasks by the shared order. Subtasks share
// their parent's category, so status, priority and due decide.
func sortSubs(items []Item, statuses []string) {
	for i := range items {
		subs := items[i].Subs
		sort.SliceStable(subs, func(a, b int) bool { return less(subs[a], subs[b], false, statuses) })
	}
}

// SortByTag orders items by their grouping tag (untagged last), then by the
// shared intra-group order, so each tag's items stay contiguous.
func SortByTag(items []Item, statuses []string) {
	sort.SliceStable(items, func(i, j int) bool {
		// catKey's "untagged/uncategorized last" trick works for tags too.
		if a, b := catKey(TagKey(items[i])), catKey(TagKey(items[j])); a != b {
			return a < b
		}
		return less(items[i], items[j], false, statuses)
	})
	sortSubs(items, statuses)
}

// SortBySource orders items by Source first (the global board view), then by
// the shared intra-group order, so each board's items stay contiguous.
func SortBySource(items []Item, statuses []string) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Source != items[j].Source {
			return items[i].Source < items[j].Source
		}
		return less(items[i], items[j], false, statuses)
	})
	sortSubs(items, statuses)
}
