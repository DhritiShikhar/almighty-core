CREATE OR REPLACE FUNCTION limit_ooe_to_space() RETURNS void as $$
-- limit_ooe_to_space() function limits order to space
	DECLARE
		i integer;
		r RECORD;
		a RECORD;
		xyz CURSOR FOR SELECT id, execution_order, space_id from work_items;
		abc CURSOR FOR SELECT id from spaces;
	BEGIN
		open abc;
			FOR a in FETCH ALL FROM abc --iterate over each space id
				LOOP
					i:=1000;
					open xyz;
						FOR r in FETCH ALL FROM xyz --fetch all workitems
							LOOP
								UPDATE work_items set execution_order=i where id=r.id AND space_id=a.id;
								i := i+1000;
							END LOOP;
					close xyz;	
				END LOOP;
		close abc;

	END $$ LANGUAGE plpgsql;

DO $$ BEGIN
	PERFORM limit_ooe_to_space();
	DROP FUNCTION limit_ooe_to_space();
END $$;
