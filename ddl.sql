-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Drop tables if they exist
DROP TABLE IF EXISTS borrowed_books;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS users;

-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create books table
CREATE TABLE books (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    published_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'Available' CHECK (status IN ('Available', 'Borrowed', 'Overdue')),
    user_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create borrowed_books table
CREATE TABLE borrowed_books (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    book_id UUID NOT NULL REFERENCES books(id),
    user_id UUID NOT NULL REFERENCES users(id),
    borrowed_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    return_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_books_status ON books(status);
CREATE INDEX idx_books_user_id ON books(user_id);
CREATE INDEX idx_borrowed_books_user_id ON borrowed_books(user_id);
CREATE INDEX idx_borrowed_books_book_id ON borrowed_books(book_id);
CREATE INDEX idx_borrowed_books_return_date ON borrowed_books(return_date);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updating updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_books_updated_at
    BEFORE UPDATE ON books
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_borrowed_books_updated_at
    BEFORE UPDATE ON borrowed_books
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample users (password is 'password123' hashed with bcrypt)
INSERT INTO users (username, password) VALUES
('john_doe', '$2a$10$xVB/FvBmT7/xDkvHl3JkaeGM5QjZHxlb0XuR1Qk3TtZZ6M7h8K9Rm'),
('jane_smith', '$2a$10$xVB/FvBmT7/xDkvHl3JkaeGM5QjZHxlb0XuR1Qk3TtZZ6M7h8K9Rm'),
('bob_wilson', '$2a$10$xVB/FvBmT7/xDkvHl3JkaeGM5QjZHxlb0XuR1Qk3TtZZ6M7h8K9Rm');

-- Insert sample books
INSERT INTO books (title, author, published_date) VALUES
('The Great Gatsby', 'F. Scott Fitzgerald', '1925-04-10T00:00:00Z'),
('To Kill a Mockingbird', 'Harper Lee', '1960-07-11T00:00:00Z'),
('1984', 'George Orwell', '1949-06-08T00:00:00Z'),
('Pride and Prejudice', 'Jane Austen', '1813-01-28T00:00:00Z'),
('The Hobbit', 'J.R.R. Tolkien', '1937-09-21T00:00:00Z'),
('The Catcher in the Rye', 'J.D. Salinger', '1951-07-16T00:00:00Z'),
('Lord of the Flies', 'William Golding', '1954-09-17T00:00:00Z'),
('Animal Farm', 'George Orwell', '1945-08-17T00:00:00Z'),
('The Little Prince', 'Antoine de Saint-Exupéry', '1943-04-06T00:00:00Z'),
('Brave New World', 'Aldous Huxley', '1932-01-01T00:00:00Z');

-- Insert some borrowed books (some returned, some not)
DO $$
DECLARE
    user1_id UUID;
    user2_id UUID;
    book1_id UUID;
    book2_id UUID;
    book3_id UUID;
BEGIN
    -- Get some user IDs
    SELECT id INTO user1_id FROM users WHERE username = 'john_doe' LIMIT 1;
    SELECT id INTO user2_id FROM users WHERE username = 'jane_smith' LIMIT 1;
    
    -- Get some book IDs
    SELECT id INTO book1_id FROM books WHERE title = 'The Great Gatsby' LIMIT 1;
    SELECT id INTO book2_id FROM books WHERE title = '1984' LIMIT 1;
    SELECT id INTO book3_id FROM books WHERE title = 'The Hobbit' LIMIT 1;

    -- Create some borrowed books
    INSERT INTO borrowed_books (book_id, user_id, borrowed_date, return_date) VALUES
    (book1_id, user1_id, CURRENT_TIMESTAMP - INTERVAL '10 days', CURRENT_TIMESTAMP - INTERVAL '3 days'),
    (book2_id, user1_id, CURRENT_TIMESTAMP - INTERVAL '5 days', NULL),
    (book3_id, user2_id, CURRENT_TIMESTAMP - INTERVAL '15 days', NULL);

    -- Update book status
    UPDATE books SET status = 'Borrowed', user_id = user1_id WHERE id = book2_id;
    UPDATE books SET status = 'Borrowed', user_id = user2_id WHERE id = book3_id;
END $$;