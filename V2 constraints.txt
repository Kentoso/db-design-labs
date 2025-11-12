-- ============================================================================
-- 1. Prevent duplicate subtypes (mutual exclusion at insert time)
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
-- 2. Campaign must have at least ONE platform
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
-- 3. Employee cannot be their own manager or mentor
-- ============================================================================

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