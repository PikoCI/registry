package migrations

var V0Initial = Migration{
	Name: "Initial",
	SQL: `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			github_id VARCHAR(255) UNIQUE,
			username VARCHAR(255) UNIQUE,
			avatar_url TEXT,
			created_at TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS namespaces (
			id VARCHAR(36) PRIMARY KEY,
			name VARCHAR(255) UNIQUE,
			type VARCHAR(50),
			github_id VARCHAR(255),
			owner_id VARCHAR(36),
			public BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP,

			CONSTRAINT fk__namespaces__users
				FOREIGN KEY (owner_id) REFERENCES users (id)
				ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS registry_types (
			id VARCHAR(36) PRIMARY KEY,
			namespace_id VARCHAR(36),
			name VARCHAR(255),
			kind VARCHAR(50),
			description TEXT,
			repository TEXT,
			homepage TEXT,
			license VARCHAR(255),
			public BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMP,

			CONSTRAINT uq__registry_types__ns_name_kind UNIQUE (namespace_id, name, kind),

			CONSTRAINT fk__registry_types__namespaces
				FOREIGN KEY (namespace_id) REFERENCES namespaces (id)
				ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS registry_versions (
			id VARCHAR(36) PRIMARY KEY,
			type_id VARCHAR(36),
			version VARCHAR(255),
			content TEXT,
			readme TEXT,
			params TEXT,
			examples TEXT,
			yanked BOOLEAN DEFAULT FALSE,
			deprecated BOOLEAN DEFAULT FALSE,
			deprecated_message TEXT,
			successor_version VARCHAR(255),
			downloads INT DEFAULT 0,
			created_at TIMESTAMP,
			created_by VARCHAR(36),

			CONSTRAINT uq__registry_versions__type_version UNIQUE (type_id, version),

			CONSTRAINT fk__registry_versions__registry_types
				FOREIGN KEY (type_id) REFERENCES registry_types (id)
				ON DELETE CASCADE,

			CONSTRAINT fk__registry_versions__users
				FOREIGN KEY (created_by) REFERENCES users (id)
		);

		CREATE TABLE IF NOT EXISTS registry_tags (
			type_id VARCHAR(36),
			tag VARCHAR(255),

			PRIMARY KEY (type_id, tag),

			CONSTRAINT fk__registry_tags__registry_types
				FOREIGN KEY (type_id) REFERENCES registry_types (id)
				ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS api_tokens (
			id VARCHAR(36) PRIMARY KEY,
			namespace_id VARCHAR(36),
			name VARCHAR(255),
			token_hash VARCHAR(255) UNIQUE,
			prefix VARCHAR(20),
			created_at TIMESTAMP,
			expires_at TIMESTAMP NULL,

			CONSTRAINT fk__api_tokens__namespaces
				FOREIGN KEY (namespace_id) REFERENCES namespaces (id)
				ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS org_members (
			org_id VARCHAR(36),
			user_id VARCHAR(36),
			role VARCHAR(50),
			created_at TIMESTAMP,

			PRIMARY KEY (org_id, user_id),

			CONSTRAINT fk__org_members__namespaces
				FOREIGN KEY (org_id) REFERENCES namespaces (id)
				ON DELETE CASCADE,

			CONSTRAINT fk__org_members__users
				FOREIGN KEY (user_id) REFERENCES users (id)
				ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS download_log (
			version_id VARCHAR(36),
			token_hash VARCHAR(255),
			ip_hash VARCHAR(255),
			download_date DATE,

			CONSTRAINT uq__download_log UNIQUE (version_id, token_hash, ip_hash, download_date),

			CONSTRAINT fk__download_log__registry_versions
				FOREIGN KEY (version_id) REFERENCES registry_versions (id)
				ON DELETE CASCADE
		);
	`,
}
