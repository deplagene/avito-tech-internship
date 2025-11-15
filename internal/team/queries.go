package team

var (
	createTeamQuery = `
		INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING;
	`

	getByNameTeamQuery = `
		SELECT user_id, username, is_active FROM users WHERE team_name = $1;
	`
)
