-- ==============================================================================
-- BIBLIO HANDS-ON SEEDER SCRIPT (VALID HEXADECIMAL UUID)
-- Purpose: Realistic mock data to explore JOIN, GROUP BY, and CTE in Beekeeper Studio
-- Note: UUID must strictly use hexadecimal characters (0-9, a-f)
--   b... = books
--   c... = copies
--   d... = loans (peminjaman)
--   e... = loan items
-- ==============================================================================

-- 1. Bersihkan data lama jika ada (agar fresh dan tidak bentrok)
TRUNCATE TABLE loan_items, loans, book_copies, books CASCADE;

-- 2. Masukkan 3 Judul Buku Induk (Catalog Books)
INSERT INTO books (id, isbn, title, author, category, published_year, created_at, updated_at) VALUES
('b1111111-1111-1111-1111-111111111111', '978-0132350884', 'Clean Code', 'Robert C. Martin', 'Software Engineering', 2008, NOW(), NOW()),
('b2222222-2222-2222-2222-222222222222', '978-0201616224', 'The Pragmatic Programmer', 'Andy Hunt', 'Software Engineering', 1999, NOW(), NOW()),
('b3333333-3333-3333-3333-333333333333', '978-1449373320', 'Designing Data-Intensive Applications', 'Martin Kleppmann', 'Distributed Systems', 2017, NOW(), NOW());

-- 3. Masukkan 6 Salinan Buku Fisik (Physical Book Copies)
INSERT INTO book_copies (id, book_id, barcode, rack_code, condition, status, created_at, updated_at) VALUES
-- Eksemplar Clean Code
('c1111111-1111-1111-1111-111111111111', 'b1111111-1111-1111-1111-111111111111', 'BC-CLN-001', 'RACK-A1', 'good', 'borrowed', NOW(), NOW()),
('c1111111-1111-1111-1111-111111111112', 'b1111111-1111-1111-1111-111111111111', 'BC-CLN-002', 'RACK-A1', 'good', 'available', NOW(), NOW()),

-- Eksemplar The Pragmatic Programmer
('c2222222-2222-2222-2222-222222222221', 'b2222222-2222-2222-2222-222222222222', 'BC-PRG-001', 'RACK-B2', 'good', 'borrowed', NOW(), NOW()),
('c2222222-2222-2222-2222-222222222222', 'b2222222-2222-2222-2222-222222222222', 'BC-PRG-002', 'RACK-B2', 'damaged', 'maintenance', NOW(), NOW()),

-- Eksemplar Designing Data-Intensive Applications
('c3333333-3333-3333-3333-333333333331', 'b3333333-3333-3333-3333-333333333333', 'BC-DDIA-001', 'RACK-C3', 'good', 'borrowed', NOW(), NOW()),
('c3333333-3333-3333-3333-333333333332', 'b3333333-3333-3333-3333-333333333333', 'BC-DDIA-002', 'RACK-C3', 'good', 'available', NOW(), NOW());

-- 4. Masukkan 4 Kasus Transaksi Peminjaman (Loans)
INSERT INTO loans (id, loan_number, member_name, status, borrowed_at, due_date, returned_at, created_at, updated_at) VALUES
-- Kasus 1: Masih dipinjam dan belum jatuh tempo (ON TIME)
('d1111111-1111-1111-1111-111111111111', 'LN-20261101-0001', 'Budi Santoso', 'borrowed', NOW() - INTERVAL '3 days', NOW() + INTERVAL '4 days', NULL, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),

-- Kasus 2: Masih dipinjam tapi SUDAH LEWAT JATUH TEMPO (OVERDUE 5 HARI!)
('d2222222-2222-2222-2222-222222222222', 'LN-20261020-0002', 'Siti Rahma', 'borrowed', NOW() - INTERVAL '12 days', NOW() - INTERVAL '5 days', NULL, NOW() - INTERVAL '12 days', NOW() - INTERVAL '12 days'),

-- Kasus 3: Sudah dikembalikan, SEMUA BUKU MULUS (NO DAMAGE)
('d3333333-3333-3333-3333-333333333333', 'LN-20261025-0003', 'Andi Wijaya', 'returned', NOW() - INTERVAL '10 days', NOW() - INTERVAL '3 days', NOW() - INTERVAL '4 days', NOW() - INTERVAL '10 days', NOW() - INTERVAL '4 days'),

-- Kasus 4: Sudah dikembalikan, TAPI ADA 1 BUKU RUSAK / SOBEK (HAS DAMAGED BOOK!)
('d4444444-4444-4444-4444-444444444444', 'LN-20261028-0004', 'Dewi Lestari', 'returned', NOW() - INTERVAL '8 days', NOW() - INTERVAL '1 days', NOW() - INTERVAL '2 days', NOW() - INTERVAL '8 days', NOW() - INTERVAL '2 days');

-- 5. Masukkan Daftar Buku yang Ada di Masing-Masing Nota (Loan Items)
INSERT INTO loan_items (id, loan_id, book_copy_id, condition_on_borrow, condition_on_return, returned_at, created_at, updated_at) VALUES
-- Item untuk Transaksi 1 (Budi pinjam 2 buku)
('e1111111-1111-1111-1111-111111111111', 'd1111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-111111111111', 'good', NULL, NULL, NOW(), NOW()),
('e1111111-1111-1111-1111-111111111112', 'd1111111-1111-1111-1111-111111111111', 'c2222222-2222-2222-2222-222222222221', 'good', NULL, NULL, NOW(), NOW()),

-- Item untuk Transaksi 2 (Siti pinjam 1 buku dan telat)
('e2222222-2222-2222-2222-222222222221', 'd2222222-2222-2222-2222-222222222222', 'c3333333-3333-3333-3333-333333333331', 'good', NULL, NULL, NOW(), NOW()),

-- Item untuk Transaksi 3 (Andi balikin 2 buku mulus semua)
('e3333333-3333-3333-3333-333333333331', 'd3333333-3333-3333-3333-333333333333', 'c1111111-1111-1111-1111-111111111112', 'good', 'good', NOW() - INTERVAL '4 days', NOW(), NOW()),
('e3333333-3333-3333-3333-333333333332', 'd3333333-3333-3333-3333-333333333333', 'c3333333-3333-3333-3333-333333333332', 'good', 'good', NOW() - INTERVAL '4 days', NOW(), NOW()),

-- Item untuk Transaksi 4 (Dewi balikin 2 buku: 1 MULUS, 1 RUSAK/DAMAGED!)
('e4444444-4444-4444-4444-444444444441', 'd4444444-4444-4444-4444-444444444444', 'c2222222-2222-2222-2222-222222222222', 'good', 'damaged', NOW() - INTERVAL '2 days', NOW(), NOW()),
('e4444444-4444-4444-4444-444444444442', 'd4444444-4444-4444-4444-444444444444', 'c1111111-1111-1111-1111-111111111111', 'good', 'good', NOW() - INTERVAL '2 days', NOW(), NOW());
