package aggregate

// AggregationParams is a narrow DTO for aggregation use-case.
type AggregationParams struct {
	MinCommentsToAnalyze int
	MaxUsersToAnalyze    int
	ExcludeAdmins        bool
}

type Params = AggregationParams

