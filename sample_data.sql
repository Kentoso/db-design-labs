BEGIN;

-- ============================================================================
-- 1. CLIENTS
-- ============================================================================
INSERT INTO client (name, email) VALUES
    ('TechCorp Solutions', 'contact@techcorp.com'),
    ('Fashion Forward Inc', 'marketing@fashionforward.com'),
    ('GreenEarth Products', 'ads@greenearth.org'),
    ('FitLife Nutrition', 'team@fitlife.com'),
    ('Urban Dwellings Real Estate', 'info@urbandwellings.com');

-- ============================================================================
-- 2. EMPLOYEES
-- ============================================================================
INSERT INTO employee (id, name, position, manager_id, mentor_id) VALUES
    (1, 'Sarah Johnson', 'CEO', NULL, NULL);

INSERT INTO employee (id, name, position, manager_id, mentor_id) VALUES
    (2, 'Michael Chen', 'VP of Marketing', 1, NULL);

-- Note: mentor_id must be UNIQUE, so each employee can have at most one mentee
INSERT INTO employee (id, name, position, manager_id, mentor_id) VALUES
    (3, 'Jessica Taylor', 'Campaign Manager', 2, 2),      -- mentored by Michael
    (4, 'Robert Martinez', 'Campaign Manager', 2, NULL),  -- no mentor assigned yet
    (5, 'Amanda Lee', 'Campaign Manager', 2, NULL),       -- no mentor assigned yet
    (6, 'Christopher Brown', 'Campaign Manager', 2, NULL); -- no mentor assigned yet

SELECT setval('employee_id_seq', 6, true);

-- ============================================================================
-- 3. AD PLATFORMS
-- ============================================================================
INSERT INTO ad_platform (name) VALUES
    ('Google Ads'),
    ('Facebook Ads'),
    ('Instagram Ads'),
    ('TikTok Ads'),
    ('LinkedIn Ads'),
    ('Twitter Ads');

-- ============================================================================
-- 4. CAMPAIGNS WITH PLATFORMS AND AD SETS
-- ============================================================================

-- Campaign 1: TechCorp
INSERT INTO campaign (id, name, start_date, finish_date, client_id, manager_id) VALUES
    (1, 'TechCorp Product Launch Q4', '2025-10-01', '2025-12-31', 1, 3);

INSERT INTO campaign_platform (campaign_id, platform_id, budget) VALUES
    (1, 1, 50000.00),  -- Google Ads
    (1, 2, 30000.00),  -- Facebook Ads
    (1, 5, 20000.00);  -- LinkedIn Ads

-- Note: ad_set names must be unique within each campaign
INSERT INTO ad_set (id, name, target_age, target_gender, target_country, campaign_id) VALUES
    (1, 'Tech Professionals 25-40', '25-40', 'all', 'US', 1),
    (2, 'Tech Enthusiasts 18-34', '18-34', 'male', 'US', 1),
    (3, 'Business Decision Makers', '35-55', 'all', 'US', 1);

-- Campaign 2: Fashion Forward
INSERT INTO campaign (id, name, start_date, finish_date, client_id, manager_id) VALUES
    (2, 'Spring Collection 2025', '2025-02-01', '2025-04-30', 2, 4);

INSERT INTO campaign_platform (campaign_id, platform_id, budget) VALUES
    (2, 2, 40000.00),  -- Facebook Ads
    (2, 3, 60000.00),  -- Instagram Ads
    (2, 4, 25000.00);  -- TikTok Ads

INSERT INTO ad_set (id, name, target_age, target_gender, target_country, campaign_id) VALUES
    (4, 'Young Women Fashion', '18-29', 'female', 'US', 2),
    (5, 'Urban Style Seekers', '25-40', 'all', 'US', 2);

-- Campaign 3: GreenEarth
INSERT INTO campaign (id, name, start_date, finish_date, client_id, manager_id) VALUES
    (3, 'Go Green Campaign 2025', '2025-01-15', '2025-06-30', 3, 5);

INSERT INTO campaign_platform (campaign_id, platform_id, budget) VALUES
    (3, 1, 35000.00),  -- Google Ads
    (3, 2, 45000.00);  -- Facebook Ads

INSERT INTO ad_set (id, name, target_age, target_gender, target_country, campaign_id) VALUES
    (6, 'Eco-Conscious Millennials', '25-40', 'all', 'US', 3),
    (7, 'Environmental Advocates', '30-55', 'all', 'US', 3);

-- Campaign 4: FitLife
INSERT INTO campaign (id, name, start_date, finish_date, client_id, manager_id) VALUES
    (4, 'FitLife Protein Launch', '2024-11-01', '2025-01-31', 4, 6);

INSERT INTO campaign_platform (campaign_id, platform_id, budget) VALUES
    (4, 3, 30000.00),  -- Instagram Ads
    (4, 4, 20000.00);  -- TikTok Ads

INSERT INTO ad_set (id, name, target_age, target_gender, target_country, campaign_id) VALUES
    (8, 'Fitness Enthusiasts', '20-35', 'all', 'US', 4),
    (9, 'Gym Goers', '25-45', 'male', 'US', 4);

-- Reset sequences
SELECT setval('campaign_id_seq', 4, true);
SELECT setval('ad_set_id_seq', 9, true);

-- ============================================================================
-- 5. MEDIA ASSETS (Videos and Images)
-- ============================================================================
-- Note: media_type field removed - type is determined by presence in video/image table

-- Videos
INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (1, 'TechCorp Product Demo', '/media/videos/techcorp_demo_v1.mp4', '2024-09-15');
INSERT INTO video (media_asset_id, duration) VALUES
    (1, 45);

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (2, 'Fashion Spring Preview', '/media/videos/fashion_spring_2025.mp4', '2025-01-10');
INSERT INTO video (media_asset_id, duration) VALUES
    (2, 30);

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (3, 'GreenEarth Story', '/media/videos/greenearth_story.mp4', '2024-12-20');
INSERT INTO video (media_asset_id, duration) VALUES
    (3, 60);

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (4, 'FitLife Workout', '/media/videos/fitlife_workout.mp4', '2024-10-25');
INSERT INTO video (media_asset_id, duration) VALUES
    (4, 15);

-- Images
INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (5, 'TechCorp Product Shot', '/media/images/techcorp_product_01.jpg', '2024-09-10');
INSERT INTO image (media_asset_id, resolution) VALUES
    (5, '1920x1080');

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (6, 'TechCorp Lifestyle', '/media/images/techcorp_lifestyle_01.jpg', '2024-09-12');
INSERT INTO image (media_asset_id, resolution) VALUES
    (6, '1200x628');

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (7, 'Fashion Model Shoot', '/media/images/fashion_model_spring_01.jpg', '2025-01-05');
INSERT INTO image (media_asset_id, resolution) VALUES
    (7, '1080x1350');

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (8, 'Fashion Collection Flat Lay', '/media/images/fashion_flatlay_01.jpg', '2025-01-08');
INSERT INTO image (media_asset_id, resolution) VALUES
    (8, '1920x1080');

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (9, 'GreenEarth Nature', '/media/images/greenearth_nature_01.jpg', '2024-12-15');
INSERT INTO image (media_asset_id, resolution) VALUES
    (9, '1920x1080');

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (10, 'GreenEarth Products', '/media/images/greenearth_products_01.jpg', '2024-12-18');
INSERT INTO image (media_asset_id, resolution) VALUES
    (10, '1200x1200');

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (11, 'FitLife Product', '/media/images/fitlife_protein_01.jpg', '2024-10-20');
INSERT INTO image (media_asset_id, resolution) VALUES
    (11, '1080x1080');

INSERT INTO media_asset (id, name, file_path, creation_date) VALUES
    (12, 'FitLife Athletes', '/media/images/fitlife_athletes_01.jpg', '2024-10-22');
INSERT INTO image (media_asset_id, resolution) VALUES
    (12, '1920x1080');

-- Reset sequence
SELECT setval('media_asset_id_seq', 12, true);

-- ============================================================================
-- 6. AD TEXTS
-- ============================================================================
-- Note: text field must be unique
INSERT INTO ad_text (id, text) VALUES
    (1, 'Introducing our revolutionary new product. Innovation meets simplicity.'),
    (2, 'Transform your workflow with TechCorp. Join 10,000+ satisfied customers.'),
    (3, 'Limited time offer: 20% off for early adopters!'),
    (4, 'Spring into style with our new collection. Fashion that makes a statement.'),
    (5, 'Sustainable fashion for the modern world. Look good, feel good.'),
    (6, 'Your wardrobe deserves an upgrade. Shop Spring 2025 now.'),
    (7, 'Join the green revolution. Every purchase plants a tree.'),
    (8, 'Eco-friendly products that don''t compromise on quality.'),
    (9, 'Mother Nature approved. Shop sustainable today.'),
    (10, 'Fuel your fitness journey. Premium protein, unbeatable taste.'),
    (11, 'From the gym to the kitchen. FitLife has you covered.'),
    (12, 'Trusted by athletes worldwide. Join the FitLife family.');

-- Reset sequence
SELECT setval('ad_text_id_seq', 12, true);

-- ============================================================================
-- 7. ADS (Complete advertisements)
-- ============================================================================

-- TechCorp Campaign Ads
INSERT INTO ad (ad_set_id, media_asset_id, ad_text_id) VALUES
    (1, 1, 1),  -- Tech Professionals: Video + Innovation text
    (1, 5, 2),  -- Tech Professionals: Product image + Customer text
    (2, 6, 3),  -- Tech Enthusiasts: Lifestyle image + Offer text
    (3, 1, 2);  -- Decision Makers: Video + Customer text

-- Fashion Forward Campaign Ads
INSERT INTO ad (ad_set_id, media_asset_id, ad_text_id) VALUES
    (4, 2, 4),  -- Young Women: Video + Statement text
    (4, 7, 5),  -- Young Women: Model image + Sustainable text
    (5, 8, 6);  -- Urban Style: Flat lay + Upgrade text

-- GreenEarth Campaign Ads
INSERT INTO ad (ad_set_id, media_asset_id, ad_text_id) VALUES
    (6, 3, 7),  -- Eco Millennials: Story video + Revolution text
    (6, 9, 8),  -- Eco Millennials: Nature image + Quality text
    (7, 10, 9); -- Environmental Advocates: Products + Approved text

-- FitLife Campaign Ads
INSERT INTO ad (ad_set_id, media_asset_id, ad_text_id) VALUES
    (8, 4, 10),  -- Fitness Enthusiasts: Workout video + Fuel text
    (8, 11, 11), -- Fitness Enthusiasts: Product image + Coverage text
    (9, 12, 12); -- Gym Goers: Athletes image + Family text

COMMIT;