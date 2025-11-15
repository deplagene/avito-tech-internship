package pullrequest

var (
	createPullRequestQuery = `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5);
	`

	getPullRequestByIdQuery = `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at,
		       ARRAY_AGG(rev.user_id) FILTER (WHERE rev.user_id IS NOT NULL) AS assigned_reviewers
		FROM pull_requests pr
		LEFT JOIN reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE pr.pull_request_id = $1
		GROUP BY pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at;
	`

	setMergeStatusQuery = `
		UPDATE pull_requests SET status = $1, merged_at = $2 WHERE pull_request_id = $3 AND status = $4;
	`

	addReviewerQuery = `
		INSERT INTO reviewers (pull_request_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;	
	`

	deleteReviewerQuery = `
		DELETE FROM reviewers WHERE pull_request_id = $1 AND user_id = $2;
	`

	getPullRequestsByReviewerQuery = `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN reviewers rev ON pr.pull_request_id = rev.pull_request_id
		WHERE rev.user_id = $1;
	`
)
