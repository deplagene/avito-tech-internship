package user

var (
	upsertUserQuery = `
		INSERT INTO users (user_id, username, team_name, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active;
	`

	getBydIdUserQuery = `
		SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1;
	`

	setIsActiveUserQuery = `
		UPDATE users SET is_active = $1 WHERE user_id = $2;	
	`

	getActiveUsersByTeamQuery = `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = TRUE AND user_id != $2
		ORDER BY RANDOM()
		LIMIT $3;	
	`

	getTeamByUserIdQuery = `
		SELECT team_name FROM users WHERE user_id = $1;	
	`
)
