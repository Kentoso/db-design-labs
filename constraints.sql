-- ============================================================================
-- MEDIA ASSET ISA INHERITANCE CONSTRAINTS
-- ============================================================================

-- ============================================================================
-- 1. Ensure media_asset has exactly ONE subtype (video XOR image)
-- ============================================================================

CREATE OR REPLACE FUNCTION check_media_asset_subtype()
RETURNS TRIGGER AS $$
DECLARE
    has_video BOOLEAN;
    has_image BOOLEAN;
BEGIN
    -- Check if this media_asset exists in video table
    SELECT EXISTS(SELECT 1 FROM video WHERE media_asset_id = NEW.id) INTO has_video;
    
    -- Check if this media_asset exists in image table
    SELECT EXISTS(SELECT 1 FROM image WHERE media_asset_id = NEW.id) INTO has_image;
    
    -- Must have exactly one subtype
    IF NOT has_video AND NOT has_image THEN
        RAISE EXCEPTION 'Media asset % must have a subtype (video or image)', NEW.id;
    END IF;
    
    IF has_video AND has_image THEN
        RAISE EXCEPTION 'Media asset % cannot be both video and image', NEW.id;
    END IF;
    
    -- Verify media_type matches the subtype
    IF has_video AND NEW.media_type != 'video' THEN
        RAISE EXCEPTION 'Media asset % has video subtype but media_type is %', NEW.id, NEW.media_type;
    END IF;
    
    IF has_image AND NEW.media_type != 'image' THEN
        RAISE EXCEPTION 'Media asset % has image subtype but media_type is %', NEW.id, NEW.media_type;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger fires AFTER INSERT/UPDATE to allow subtype to be created first
CREATE CONSTRAINT TRIGGER enforce_media_asset_subtype
    AFTER INSERT OR UPDATE ON media_asset
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW
    EXECUTE FUNCTION check_media_asset_subtype();

-- ============================================================================
-- 2. Ensure video/image media_type matches
-- ============================================================================

CREATE OR REPLACE FUNCTION check_video_media_type()
RETURNS TRIGGER AS $$
DECLARE
    actual_media_type TEXT;
BEGIN
    SELECT media_type INTO actual_media_type 
    FROM media_asset 
    WHERE id = NEW.media_asset_id;
    
    IF actual_media_type IS NULL THEN
        RAISE EXCEPTION 'Media asset % does not exist', NEW.media_asset_id;
    END IF;
    
    IF actual_media_type != 'video' THEN
        RAISE EXCEPTION 'Media asset % has media_type %, expected video', 
            NEW.media_asset_id, actual_media_type;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_video_media_type
    BEFORE INSERT OR UPDATE ON video
    FOR EACH ROW
    EXECUTE FUNCTION check_video_media_type();

CREATE OR REPLACE FUNCTION check_image_media_type()
RETURNS TRIGGER AS $$
DECLARE
    actual_media_type TEXT;
BEGIN
    SELECT media_type INTO actual_media_type 
    FROM media_asset 
    WHERE id = NEW.media_asset_id;
    
    IF actual_media_type IS NULL THEN
        RAISE EXCEPTION 'Media asset % does not exist', NEW.media_asset_id;
    END IF;
    
    IF actual_media_type != 'image' THEN
        RAISE EXCEPTION 'Media asset % has media_type %, expected image', 
            NEW.media_asset_id, actual_media_type;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_image_media_type
    BEFORE INSERT OR UPDATE ON image
    FOR EACH ROW
    EXECUTE FUNCTION check_image_media_type();


-- ============================================================================
-- 3. Prevent duplicate subtypes
-- ============================================================================

CREATE OR REPLACE FUNCTION prevent_duplicate_subtype()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_TABLE_NAME = 'video' THEN
        IF EXISTS(SELECT 1 FROM image WHERE media_asset_id = NEW.media_asset_id) THEN
            RAISE EXCEPTION 'Media asset % already exists as image, cannot create video', 
                NEW.media_asset_id;
        END IF;
    ELSIF TG_TABLE_NAME = 'image' THEN
        IF EXISTS(SELECT 1 FROM video WHERE media_asset_id = NEW.media_asset_id) THEN
            RAISE EXCEPTION 'Media asset % already exists as video, cannot create image', 
                NEW.media_asset_id;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_video_image_duplicate
    BEFORE INSERT ON video
    FOR EACH ROW
    EXECUTE FUNCTION prevent_duplicate_subtype();

CREATE TRIGGER prevent_image_video_duplicate
    BEFORE INSERT ON image
    FOR EACH ROW
    EXECUTE FUNCTION prevent_duplicate_subtype();


-- ============================================================================
-- MINIMUM CARDINALITY CONSTRAINTS
-- ============================================================================

-- ============================================================================
-- 1. Campaign must have at least ONE platform
-- ============================================================================

CREATE OR REPLACE FUNCTION check_campaign_has_platform()
RETURNS TRIGGER AS $$
DECLARE
    platform_count INTEGER;
BEGIN
    -- Count remaining platforms for this campaign after the delete
    SELECT COUNT(*) INTO platform_count
    FROM campaign_platform
    WHERE campaign_id = OLD.campaign_id
      AND (campaign_id != OLD.campaign_id OR platform_id != OLD.platform_id);
    
    IF platform_count = 0 THEN
        RAISE EXCEPTION 'Cannot delete last platform for campaign %. Campaign must have at least one platform.', 
            OLD.campaign_id;
    END IF;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_campaign_min_platforms
    BEFORE DELETE ON campaign_platform
    FOR EACH ROW
    EXECUTE FUNCTION check_campaign_has_platform();

-- Verify on campaign INSERT/UPDATE that it has at least one platform (deferred)
CREATE OR REPLACE FUNCTION check_campaign_platform_exists()
RETURNS TRIGGER AS $$
DECLARE
    platform_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO platform_count
    FROM campaign_platform
    WHERE campaign_id = NEW.id;
    
    IF platform_count = 0 THEN
        RAISE EXCEPTION 'Campaign % must have at least one platform', NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER enforce_campaign_has_platforms
    AFTER INSERT OR UPDATE ON campaign
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW
    EXECUTE FUNCTION check_campaign_platform_exists();


-- ============================================================================
-- 2. Campaign must have at least ONE ad_set
-- ============================================================================

CREATE OR REPLACE FUNCTION check_campaign_has_ad_set()
RETURNS TRIGGER AS $$
DECLARE
    ad_set_count INTEGER;
BEGIN
    -- Count remaining ad_sets for this campaign after the delete
    SELECT COUNT(*) INTO ad_set_count
    FROM ad_set
    WHERE campaign_id = OLD.campaign_id
      AND id != OLD.id;
    
    IF ad_set_count = 0 THEN
        RAISE EXCEPTION 'Cannot delete last ad_set for campaign %. Campaign must have at least one ad_set.', 
            OLD.campaign_id;
    END IF;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_campaign_min_ad_sets
    BEFORE DELETE ON ad_set
    FOR EACH ROW
    EXECUTE FUNCTION check_campaign_has_ad_set();

-- Verify on campaign INSERT/UPDATE that it has at least one ad_set (deferred)
CREATE OR REPLACE FUNCTION check_campaign_ad_set_exists()
RETURNS TRIGGER AS $$
DECLARE
    ad_set_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO ad_set_count
    FROM ad_set
    WHERE campaign_id = NEW.id;
    
    IF ad_set_count = 0 THEN
        RAISE EXCEPTION 'Campaign % must have at least one ad_set', NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER enforce_campaign_has_ad_sets
    AFTER INSERT OR UPDATE ON campaign
    DEFERRABLE INITIALLY DEFERRED
    FOR EACH ROW
    EXECUTE FUNCTION check_campaign_ad_set_exists();

CREATE OR REPLACE FUNCTION check_employee_self_reference()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.manager_id IS NOT NULL AND NEW.manager_id = NEW.id THEN
        RAISE EXCEPTION 'Employee % cannot be their own manager', NEW.id;
    END IF;

    IF NEW.mentor_id IS NOT NULL AND NEW.mentor_id = NEW.id THEN
        RAISE EXCEPTION 'Employee % cannot be their own mentor', NEW.id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_employee_self_reference
    BEFORE INSERT OR UPDATE ON employee
    FOR EACH ROW
    EXECUTE FUNCTION check_employee_self_reference();