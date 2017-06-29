ALTER TABLE profiles ADD tags text;

DO
$$
DECLARE
    _profile RECORD;
    _tags text;
    _tag RECORD;
BEGIN
    RAISE NOTICE 'Converting tags from tags table back to profile table...';
    FOR _profile IN SELECT * FROM profiles ORDER BY profile_id LOOP
    	_tags := '';
    	FOR _tag IN SELECT * FROM tags WHERE profile_id=_profile.profile_id LOOP
        	_tags = _tags || ',' || _tag.tag;
        END LOOP;
        _tags = ltrim(_tags, ',');
        RAISE NOTICE 'Tags: %', _tags;
        UPDATE profiles SET tags=_tags WHERE profile_id=_profile.profile_id;
    END LOOP;
    RAISE NOTICE 'Done converting tags.';
END;
$$
;
DROP TABLE tags;