DELETE FROM journal_entries WHERE source_type='expense' AND source_id IN (
  '389a5e77-985c-4cb1-8be2-b16c254080b3',
  '7e06266c-a895-4b42-9049-defc62176e6d',
  '245eccdd-dc5c-4083-ab37-3087fc6c3851',
  '447dcdd7-bf3f-46cd-a709-f649cbd6efa0',
  '7e817fbb-c39e-4be6-b5ac-79774db864bc',
  'afbf2275-555d-4673-b99d-6a59680dde17',
  '2ac23cd1-e62c-4bdb-935e-4365a702deff',
  '0187d429-3f4b-4b5a-9675-f580ad1462a7',
  '0fc33c2c-8e08-4fb9-baf5-fdb97460ce82',
  '91ef8400-84e8-433b-98dc-0e89cffcebf3',
  'c5fdd8e7-d6b0-48c9-a846-5bc63ae3059f',
  '9d330096-7e4c-4160-a8b1-1f84b3faa23c',
  '8b83b93c-e92f-416b-990e-e0cda3c6e906',
  '56607835-25c9-49cf-9f64-791ae0735187',
  '9bafd483-c30a-42a7-a679-2cf6af172eb5',
  '2d9cb21e-3cd6-45d2-8f38-31fd65dfd068'
);

DELETE FROM expenses WHERE id IN (
  '389a5e77-985c-4cb1-8be2-b16c254080b3',
  '7e06266c-a895-4b42-9049-defc62176e6d',
  '245eccdd-dc5c-4083-ab37-3087fc6c3851',
  '447dcdd7-bf3f-46cd-a709-f649cbd6efa0',
  '7e817fbb-c39e-4be6-b5ac-79774db864bc',
  'afbf2275-555d-4673-b99d-6a59680dde17',
  '2ac23cd1-e62c-4bdb-935e-4365a702deff',
  '0187d429-3f4b-4b5a-9675-f580ad1462a7',
  '0fc33c2c-8e08-4fb9-baf5-fdb97460ce82',
  '91ef8400-84e8-433b-98dc-0e89cffcebf3',
  'c5fdd8e7-d6b0-48c9-a846-5bc63ae3059f',
  '9d330096-7e4c-4160-a8b1-1f84b3faa23c',
  '8b83b93c-e92f-416b-990e-e0cda3c6e906',
  '56607835-25c9-49cf-9f64-791ae0735187',
  '9bafd483-c30a-42a7-a679-2cf6af172eb5',
  '2d9cb21e-3cd6-45d2-8f38-31fd65dfd068'
);
