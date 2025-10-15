DROP TABLE IF EXISTS ad CASCADE;
DROP TABLE IF EXISTS ad_text CASCADE;
DROP TABLE IF EXISTS image CASCADE;
DROP TABLE IF EXISTS video CASCADE;
DROP TABLE IF EXISTS media_asset CASCADE;
DROP TABLE IF EXISTS ad_set CASCADE;
DROP TABLE IF EXISTS campaign_platform CASCADE;
DROP TABLE IF EXISTS ad_platform CASCADE;
DROP TABLE IF EXISTS campaign CASCADE;
DROP TABLE IF EXISTS employee CASCADE;
DROP TABLE IF EXISTS client CASCADE;

CREATE TABLE client (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE employee (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    position TEXT NOT NULL,
    manager_id INTEGER REFERENCES employee(id) ON DELETE SET NULL,
    mentor_id INTEGER REFERENCES employee(id) ON DELETE SET NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE campaign (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    start_date DATE NOT NULL,
    finish_date DATE NOT NULL,
    client_id INTEGER NOT NULL REFERENCES client(id) ON DELETE CASCADE,
    manager_id INTEGER NOT NULL REFERENCES employee(id) ON DELETE RESTRICT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_dates CHECK (finish_date >= start_date)
);

CREATE TABLE ad_platform (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE campaign_platform (
    campaign_id INTEGER NOT NULL REFERENCES campaign(id) ON DELETE CASCADE,
    platform_id INTEGER NOT NULL REFERENCES ad_platform(id) ON DELETE CASCADE,
    budget DECIMAL(12, 2) NOT NULL CHECK (budget >= 0),
    PRIMARY KEY (campaign_id, platform_id)
);

CREATE TABLE ad_set (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    target_age TEXT,
    target_gender TEXT,
    target_country TEXT,
    campaign_id INTEGER NOT NULL REFERENCES campaign(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (campaign_id, name)
);

CREATE TABLE media_asset (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    file_path TEXT NOT NULL UNIQUE,
    creation_date DATE NOT NULL DEFAULT CURRENT_DATE
);

CREATE TABLE video (
    media_asset_id INTEGER PRIMARY KEY REFERENCES media_asset(id) ON DELETE CASCADE,
    duration INTEGER NOT NULL CHECK (duration > 0 AND duration <= 600) -- 10 minutes max
);

CREATE TABLE image (
    media_asset_id INTEGER PRIMARY KEY REFERENCES media_asset(id) ON DELETE CASCADE,
    resolution TEXT NOT NULL
);

CREATE TABLE ad_text (
    id SERIAL PRIMARY KEY,
    text TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ad (
    id SERIAL PRIMARY KEY,
    ad_set_id INTEGER NOT NULL REFERENCES ad_set(id) ON DELETE CASCADE,
    media_asset_id INTEGER NOT NULL REFERENCES media_asset(id) ON DELETE CASCADE,
    ad_text_id INTEGER NOT NULL REFERENCES ad_text(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (ad_set_id, media_asset_id, ad_text_id)
);