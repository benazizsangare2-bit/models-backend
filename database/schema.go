package database

func GetTableQueries() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
    		userid SERIAL PRIMARY KEY,
    		email TEXT UNIQUE NOT NULL,
    		email_verified BOOLEAN DEFAULT FALSE,
    		verification_code TEXT,             -- OTP code
    		verification_expiry TIMESTAMP WITH TIME ZONE,      -- When the OTP expires
    		password TEXT,                      -- stored after verification
    		fullname TEXT,
    		username TEXT,
    		phone_number TEXT UNIQUE,
    		position TEXT,
    		created_at TIMESTAMP DEFAULT NOW()
);`,
		`CREATE TABLE IF NOT EXISTS models (
		    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id INT REFERENCES 	users(userid) not null,
		    first_name VARCHAR(100) NOT NULL,
		    last_name VARCHAR(100) NOT NULL,
		    username VARCHAR(50) UNIQUE NOT NULL,
		    email VARCHAR(255) UNIQUE NOT NULL,
		    whatsapp VARCHAR(20) NOT NULL,
		    date_of_birth DATE NOT NULL,
		    gender VARCHAR(10) NOT NULL CHECK (gender IN ('Female','Male','Other')),
		    nationality VARCHAR(100) NOT NULL,
			street VARCHAR(100) NOT NULL,
			city VARCHAR(100) NOT NULL,
		    residence_country VARCHAR(100) NOT NULL,
		    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending','under_review','approved','rejected')),
		    deleted BOOLEAN DEFAULT FALSE,
		    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			registration_step INT DEFAULT 1
);
`, `CREATE TABLE IF NOT EXISTS model_measurements (
   			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   			model_id UUID REFERENCES models(id) NOT NULL,
   			experience VARCHAR(100) NOT NULL,
   			height INT NOT NULL,
   			weight INT NOT NULL,
   			waist INT NOT NULL,
   			hips INT NOT NULL,
   			hair_color VARCHAR(100) NOT NULL,
   			eye_color VARCHAR(100) NOT NULL,
   			photo TEXT,
   			additional_photos TEXT 
);
`, `CREATE TABLE IF NOT EXISTS model_documents (
    		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    		model_id UUID NOT NULL REFERENCES models(id) ON DELETE CASCADE,
    		document_issuer_country VARCHAR(100) NOT NULL,
    		document_type VARCHAR(50) NOT NULL CHECK (document_type IN ('National ID Card', 'Passport', 'Driver''s License')),
    		document_front TEXT NOT NULL,  -- will store the file path or URL
    		document_back TEXT NOT NULL,   -- same
    		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`, `CREATE TABLE IF NOT EXISTS model_identity_check (
    		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    		model_id UUID REFERENCES models(id) ON DELETE CASCADE,
    		selfie_with_id TEXT NOT NULL, -- file path
    		verified BOOLEAN DEFAULT FALSE,
    		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


-- Hostesses Schema here


`, `CREATE TABLE IF NOT EXISTS hostesses (
    	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    	user_id INT REFERENCES users(userid) NOT NULL,

    	-- Personal info
    	first_name VARCHAR(100) NOT NULL,
    	last_name VARCHAR(100) NOT NULL,
    	username VARCHAR(50) UNIQUE NOT NULL,
    	email VARCHAR(255) UNIQUE NOT NULL,
    	whatsapp VARCHAR(20) NOT NULL,
    	date_of_birth DATE NOT NULL,
    	gender VARCHAR(10) NOT NULL CHECK (gender IN ('Female','Male','Other')),
    	nationality VARCHAR(100) NOT NULL,
    	street VARCHAR(100) NOT NULL,
    	city VARCHAR(100) NOT NULL,
    	residence_country VARCHAR(100) NOT NULL,

		-- Status tracking
    	registration_step INT DEFAULT 1,
    	status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending','under_review','approved','rejected')),
		deleted BOOLEAN DEFAULT FALSE,

		emergency_contact_name VARCHAR(100),
    	emergency_contact_relationship VARCHAR(50),
    	emergency_contact_phone VARCHAR(20),


    	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

`, `CREATE TABLE IF NOT EXISTS hostess_experience (
    	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    	hostess_id UUID REFERENCES hostesses(id) ON DELETE CASCADE,

    	work_experience TEXT,
    	languages TEXT[],        -- e.g. ['English','French','Spanish']
    	skills TEXT[],            -- e.g. ['Communication','Customer Service']
    	availability VARCHAR(100),
    	preferred_events TEXT[],  -- e.g. ['Conferences','Fashion Shows']
    	previous_hostess_work TEXT,
		reference_contact TEXT,

		height VARCHAR(10),
    	weight VARCHAR(10),
    	hair_color VARCHAR(50),
    	eye_color VARCHAR(50),

    	-- Media
    	photo TEXT,
    	additional_photos TEXT[],

    	-- Social links
    	social_instagram TEXT,
    	social_facebook TEXT,
    	social_twitter TEXT,
    	social_linkedin TEXT,

		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()

);
`, `CREATE TABLE IF NOT EXISTS hostess_documents (
    	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    	hostess_id UUID NOT NULL REFERENCES hostesses(id) ON DELETE CASCADE,
    	document_issuer_country VARCHAR(100) NOT NULL,
    	document_type VARCHAR(50) NOT NULL CHECK (
    	    document_type IN ('National ID Card', 'Passport', 'Driver''s License')
    	),
    	document_front TEXT NOT NULL,  -- file path or URL
    	document_back TEXT NOT NULL,   -- file path or URL
    	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    	updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`, `CREATE TABLE IF NOT EXISTS hostess_identity_check (
    	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    	hostess_id UUID NOT NULL REFERENCES hostesses(id) ON DELETE CASCADE,
    	selfie_with_id TEXT NOT NULL,  -- selfie file path
    	verified BOOLEAN DEFAULT FALSE,
    	created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`,
	}
}
