CREATE TABLE tags (
    profile_id bigint,
    tag text
);

DO
$$
DECLARE
    _profile RECORD;
    _tags text[];
    _tag text;
BEGIN
    RAISE NOTICE 'Converting tags from profile table to new tags table...';
    FOR _profile IN SELECT * FROM profiles ORDER BY profile_id LOOP
    	RAISE NOTICE 'Tags: %', _profile.tags;
        IF _profile.tags != '' THEN
        	_tags := regexp_split_to_array(_profile.tags, ',');
        
            RAISE NOTICE 'Tags length: %', array_length(_tags,1);
            IF array_length(_tags,1) != 0 THEN
                FOREACH _tag IN ARRAY _tags
                LOOP
                    RAISE NOTICE 'Tag: %', _tag;
                    _tag = ltrim(_tag, ' ');
                    _tag = rtrim(_tag, ' ');
                    IF _tag != '' THEN
                    	INSERT INTO tags (profile_id, tag) VALUES (_profile.profile_id, _tag);
                    END IF;
                END LOOP;
            END IF;
        END IF;
    END LOOP;
    RAISE NOTICE 'Done converting tags.';
END;
$$
;
ALTER TABLE profiles DROP IF EXISTS tags;