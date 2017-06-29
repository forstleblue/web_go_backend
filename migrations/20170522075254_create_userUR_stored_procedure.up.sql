CREATE TYPE ur_feedback AS (
    ur_score SMALLINT,
    ur_rating_count INT,
    sda_val TEXT[]
);

CREATE OR REPLACE FUNCTION get_user_ur(BIGINT) RETURNS ur_feedback AS 
$$

    DECLARE 
        result_record ur_feedback;
        inp_user_id ALIAS FOR $1;
        val TEXT;
        i INTEGER :=1;
    BEGIN
        SELECT coalesce(avg(score), 0)
        INTO result_record.ur_score
        FROM feedbacks
        WHERE receiver_profile_id IN (SELECT profile_id FROM profiles WHERE user_id = inp_user_id);

        SELECT COUNT(score)
        INTO result_record.ur_rating_count
        FROM feedbacks
        WHERE receiver_profile_id IN (SELECT profile_id FROM profiles WHERE user_id = inp_user_id);

        FOR val IN SELECT sda FROM (SELECT UNNEST(sda_text) AS sda FROM (SELECT * FROM feedbacks WHERE receiver_profile_id IN (SELECT profile_id FROM profiles WHERE user_id = inp_user_id)) profile_feedback GROUP BY sda ORDER BY COUNT(UNNEST(sda_text)) DESC LIMIT 3) sda_list LOOP
            result_record.sda_val[i] = val;
            i := i + 1;
        END LOOP; 

        RETURN result_record;

    END
$$ LANGUAGE plpgsql;