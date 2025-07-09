import { writable } from 'svelte/store';

export interface Book {
  id: string;
  title: string;
  author: string;
  coverUrl?: string;
  epubUrl: string; // Path to the EPUB file, relative to where they are served
  progress?: number; // Percentage, 0-100
  // Add other relevant fields like description, seriesInfo, etc. later
}

// Mock Data
const mockBooks: Book[] = [
  {
    id: '1',
    title: 'The Great Gatsby',
    author: 'F. Scott Fitzgerald',
    coverUrl: 'https://covers.openlibrary.org/b/id/8264810-L.jpg', // Example cover
    epubUrl: '/books/dummy.epub', // Assuming dummy.epub can be used for all for now
    progress: 25,
  },
  {
    id: '2',
    title: 'To Kill a Mockingbird',
    author: 'Harper Lee',
    coverUrl: 'https://covers.openlibrary.org/b/id/10747280-L.jpg', // Example cover
    epubUrl: '/books/dummy.epub',
    progress: 60,
  },
  {
    id: '3',
    title: '1984',
    author: 'George Orwell',
    // No cover to test fallback
    epubUrl: '/books/dummy.epub',
    progress: 0,
  },
  {
    id: '4',
    title: 'Pride and Prejudice',
    author: 'Jane Austen',
    coverUrl: 'https://covers.openlibrary.org/b/id/1003918-L.jpg', // Example cover
    epubUrl: '/books/dummy.epub',
  },
];

export const library = writable<Book[]>(mockBooks);

// Function to update progress (example)
export function updateBookProgress(bookId: string, newProgress: number) {
  library.update(books =>
    books.map(book =>
      book.id === bookId ? { ...book, progress: Math.max(0, Math.min(100, newProgress)) } : book
    )
  );
}

// Example: Simulate progress update after some time for the first book
setTimeout(() => {
    updateBookProgress('1', 35);
}, 15000);
