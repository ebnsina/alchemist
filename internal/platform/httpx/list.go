package httpx

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// The list contract every collection endpoint shares: limit, offset, q, sort, order.
// It lives here rather than in api/ so a module can serve the same contract without
// importing the control plane.

// ListMax is the most rows one page may ask for. Beyond it a caller is paging by
// accident; a database that has to sort a whole library per request is the cost.
const ListMax = 100

// List is one page of a collection.
type List struct {
	Limit  int
	Offset int
	// Q is the free-text term, already trimmed. Empty means no search.
	Q     string
	sort  string
	order string
}

// OrderBy is safe to interpolate into SQL: sort came from the endpoint's own
// allow-list and order is one of two literals, so no caller text reaches the query.
func (l List) OrderBy() string { return l.sort + " " + l.order }

// ParseList reads the shared parameters, or writes a 400 and returns false.
//
// It refuses rather than clamps on purpose: the frontend maps stable codes to
// messages, and quietly returning a different page than the one asked for is worse
// than a refusal somebody can see.
func ParseList(w http.ResponseWriter, r *http.Request, sortable []string, defaultSort string) (List, bool) {
	q := r.URL.Query()
	l := List{Limit: 25, Q: strings.TrimSpace(q.Get("q")), sort: defaultSort, order: "desc"}

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > ListMax {
			ErrorFor(w, r, http.StatusBadRequest, "invalid_limit",
				"Ask for between 1 and "+strconv.Itoa(ListMax)+" at a time.")
			return l, false
		}
		l.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			ErrorFor(w, r, http.StatusBadRequest, "invalid_offset",
				"Start from 0 or a higher whole number.")
			return l, false
		}
		l.Offset = n
	}
	if v := q.Get("sort"); v != "" {
		if !slices.Contains(sortable, v) {
			ErrorFor(w, r, http.StatusBadRequest, "invalid_sort",
				"You can sort this list by "+strings.Join(sortable, ", ")+".")
			return l, false
		}
		l.sort = v
	}
	if v := q.Get("order"); v != "" {
		if v != "asc" && v != "desc" {
			ErrorFor(w, r, http.StatusBadRequest, "invalid_order",
				"Order by asc or desc.")
			return l, false
		}
		l.order = v
	}
	return l, true
}

// Flag reads a filter that is either true, false, or absent. Anything else is a
// refusal: "revoked=maybe" silently meaning "all" hides rows the caller asked about.
func Flag(w http.ResponseWriter, r *http.Request, name string) (*bool, bool) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return nil, true
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		ErrorFor(w, r, http.StatusBadRequest, "invalid_filter",
			"Set "+name+" to true or false.")
		return nil, false
	}
	return &b, true
}
