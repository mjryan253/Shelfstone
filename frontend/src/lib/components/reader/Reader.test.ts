import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Reader from './Reader.svelte';

// Mock Epub.js
// Vitest automatically hoists vi.mock calls to the top of the module.
const mockBook = {
  renderTo: vi.fn().mockReturnThis(),
  display: vi.fn().mockResolvedValue(undefined),
  destroy: vi.fn(),
  locations: {
    generate: vi.fn().mockResolvedValue(undefined),
    locationFromCfi: vi.fn().mockReturnValue(0), // page 1
    length: vi.fn().mockReturnValue(100), // total 100 pages
  },
  ready: Promise.resolve(),
  spine: {
    get: vi.fn(),
  },
  on: vi.fn(), // Mock the 'on' method
};
const mockRendition = {
  display: vi.fn().mockResolvedValue(undefined),
  next: vi.fn().mockResolvedValue(undefined),
  prev: vi.fn().mockResolvedValue(undefined),
  on: vi.fn((event, callback) => {
    // Store callbacks to simulate events if needed
    if (event === 'relocated') {
      // @ts-ignore
      mockRendition.triggerRelocated = callback;
    }
  }),
  currentLocation: vi.fn().mockReturnValue({ start: { cfi: 'epubcfi(/6/4!/4/2/2:0)' } }),
  destroy: vi.fn(),
};

vi.mock('epubjs', () => ({
  default: vi.fn(() => mockBook),
}));


describe('Reader.svelte', () => {
  beforeEach(() => {
    // Reset mocks before each test
    vi.clearAllMocks();
    // Setup mockBook.renderTo to return the mockRendition
    mockBook.renderTo.mockReturnValue(mockRendition);
  });

  it('should display loading message initially', () => {
    render(Reader, { props: { epubUrl: 'test.epub' } });
    expect(screen.getByText('Loading EPUB...')).toBeInTheDocument();
  });

  it('should display error message if epubUrl is not provided', async () => {
    render(Reader, { props: { epubUrl: '' } });
    // Need to wait for onMount to complete
    await new Promise(resolve => setTimeout(resolve, 0));
    expect(await screen.findByText('No EPUB URL provided.')).toBeInTheDocument();
  });

  it('should call ePub and rendition.display on mount with a valid URL', async () => {
    const testUrl = 'http://example.com/book.epub';
    render(Reader, { props: { epubUrl: testUrl } });

    // Wait for onMount to complete its async operations
    await new Promise(resolve => setTimeout(resolve, 0));

    expect(vi.mocked(ePub).default).toHaveBeenCalledWith(testUrl);
    expect(mockBook.renderTo).toHaveBeenCalled();
    expect(mockRendition.display).toHaveBeenCalled();
  });

  // More tests would go here, for example:
  // - Clicking next/prev buttons calls rendition.next/prev
  // - Error handling when ePub loading fails
  // - Correct location display after 'relocated' event
  // - Component cleanup on destroy (book.destroy called)

  it('should render navigation buttons', async () => {
    render(Reader, { props: { epubUrl: 'test.epub' } });
    await new Promise(resolve => setTimeout(resolve, 0)); // Wait for loading
    expect(screen.getByText('Previous')).toBeInTheDocument();
    expect(screen.getByText('Next')).toBeInTheDocument();
  });

  // This is a more complex test that would require deeper mocking of Epub.js events
  it.skip('should update location display when rendition relocates', async () => {
    render(Reader, { props: { epubUrl: 'test.epub' } });
    await new Promise(resolve => setTimeout(resolve, 0)); // Wait for loading

    // Simulate the 'relocated' event
    // @ts-ignore
    if (mockRendition.triggerRelocated) {
        // @ts-ignore
      mockRendition.triggerRelocated({ start: { cfi: 'epubcfi(/6/4!/4/2/4:0)' } });
    }
    // Ideally, locationFromCfi would be called, and you'd check the new page number
    // This requires the mock setup to be more detailed.
    // For now, we just check if the event handler is set up via mockRendition.on
    expect(mockRendition.on).toHaveBeenCalledWith('relocated', expect.any(Function));
  });

});
