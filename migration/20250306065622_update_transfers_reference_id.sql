-- +goose Up
-- +goose StatementBegin
UPDATE transfers t 
JOIN payments p 
ON p.reference_id = SUBSTRING_INDEX(t.reference_id, "/SPLIT/", 1)
SET t.reference_id = p.uuid 
WHERE t.reference_id LIKE "%/SPLIT/%";
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE transfers
SET reference_id = CASE 
                  WHEN uuid = "0fe149a4-beab-40d1-b6d3-4a1463809599" THEN 'QA202502040004/SPLIT/0001'
                  WHEN uuid = "5dbf63e9-d430-4d92-9ea9-c30a82ad3088" THEN '1740582343/SPLIT/0001'
                  WHEN uuid = "61ace24e-1266-4fb9-a27c-66910f3f9391" THEN 'QA202502040004/SPLIT/0002'
              END
WHERE uuid IN ("0fe149a4-beab-40d1-b6d3-4a1463809599", "5dbf63e9-d430-4d92-9ea9-c30a82ad3088", "61ace24e-1266-4fb9-a27c-66910f3f9391");
-- +goose StatementEnd
