import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { library, updateBookProgress, type Book } from './library';

// Helper to reset store to initial mock state before each test
const getInitialMockBooks = (): Book[] => [
  {
    id: '1',
    title: 'The Great Gatsby',
    author: 'F. Scott Fitzgerald',
    coverUrl: 'https://covers.openlibrary.org/b/id/8264810-L.jpg',
    epubUrl: '/books/dummy.epub',
    progress: 25,
  },
  {
    id: '2',
    title: 'To Kill a Mockingbird',
    author: 'Harper Lee',
    coverUrl: 'https://covers.openlibrary.org/b/id/10747280-L.jpg',
    epubUrl: '/books/dummy.epub',
    progress: 60,
  },
  {
    id: '3',
    title: '1984',
    author: 'George Orwell',
    epubUrl: '/books/dummy.epub',
    progress: 0,
  },
  {
    id: '4',
    title: 'Pride and Prejudice',
    author: 'Jane Austen',
    coverUrl: 'https://covers.openlibrary.org/b/id/1003918-L.jpg',
    epubUrl: '/books/dummy.epub',
    // progress undefined
  },
];

describe('library store', () => {
  beforeEach(() => {
    // Reset the store to a known state before each test
    library.set(getInitialMockBooks());
  });

  it('should initialize with mock books', () => {
    const books = get(library);
    expect(books.length).toBe(4);
    expect(books[0].title).toBe('The Great Gatsby');
  });

  it('updateBookProgress should update the progress of the correct book', () => {
    updateBookProgress('1', 50);
    const books = get(library);
    const updatedBook = books.find(b => b.id === '1');
    expect(updatedBook?.progress).toBe(50);
  });

  it('updateBookProgress should not affect other books', () => {
    updateBookProgress('1', 50);
    const books = get(library);
    const otherBook = books.find(b => b.id === '2');
    expect(otherBook?.progress).toBe(60); // Its original progress
  });

  it('updateBookProgress should clamp progress to 0-100 range (over)', () => {
    updateBookProgress('1', 150);
    const books = get(library);
    const updatedBook = books.find(b => b.id === '1');
    expect(updatedBook?.progress).toBe(100);
  });

  it('updateBookProgress should clamp progress to 0-100 range (under)', () => {
    updateBookProgress('1', -50);
    const books = get(library);
    const updatedBook = books.find(b => b.id === '1');
    expect(updatedBook?.progress).toBe(0);
  });

  it('updateBookProgress should handle book not found gracefully', () => {
    const initialBooks = get(library);
    updateBookProgress('nonexistent-id', 50);
    const currentBooks = get(library);
    expect(currentBooks).toEqual(initialBooks);
  });

  it('updateBookProgress should correctly set progress for a book that had undefined progress', () => {
    updateBookProgress('4', 75);
    const books = get(library);
    const updatedBook = books.find(b => b.id === '4');
    expect(updatedBook?.progress).toBe(75);
  });
});
