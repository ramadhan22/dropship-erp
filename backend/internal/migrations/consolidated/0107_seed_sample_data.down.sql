-- Clear sample data
DELETE FROM journal_lines WHERE memo LIKE 'Sample%';
DELETE FROM journal_entries WHERE description LIKE 'Sample%';
DELETE FROM expense_lines WHERE description LIKE '%purchase';
DELETE FROM expenses WHERE description LIKE 'Office Supplies%' OR description LIKE 'Marketing Materials%';