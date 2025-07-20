# 📦 ShopGPT Frontend Architecture

## 🏛️ Architecture Overview

ShopGPT uses a **Feature-Sliced Design (FSD)** combined with **Atomic Design** principles to create a scalable, maintainable, and testable frontend architecture.

## 📁 Folder Structure

```
apps/web/
├── app/                          # Next.js App Router
│   ├── (auth)/                  # Auth group routes
│   │   ├── login/
│   │   └── register/
│   ├── (chat)/                  # Chat group routes
│   │   ├── layout.tsx
│   │   └── chat/
│   │       ├── page.tsx
│   │       └── [id]/page.tsx
│   ├── api/                     # API Routes
│   ├── layout.tsx               # Root layout
│   ├── page.tsx                 # Home page
│   └── globals.css
│
├── src/
│   ├── app/                     # App-wide configurations
│   │   ├── providers/           # React context providers
│   │   │   ├── ThemeProvider.tsx
│   │   │   ├── AuthProvider.tsx
│   │   │   └── StoreProvider.tsx
│   │   ├── styles/              # Global styles
│   │   └── config/              # App configuration
│   │
│   ├── processes/               # Business processes
│   │   ├── auth/                # Authentication flow
│   │   ├── checkout/            # Purchase flow
│   │   └── search/              # Search flow
│   │
│   ├── pages/                   # Page components
│   │   ├── ChatPage/
│   │   ├── HomePage/
│   │   └── ProfilePage/
│   │
│   ├── widgets/                 # Large UI blocks
│   │   ├── Header/
│   │   ├── Sidebar/
│   │   ├── ChatArea/
│   │   └── ProductGrid/
│   │
│   ├── features/                # Feature-specific logic
│   │   ├── chat/                # Chat feature
│   │   │   ├── ui/
│   │   │   ├── model/
│   │   │   ├── api/
│   │   │   └── lib/
│   │   ├── search/              # Search feature
│   │   │   ├── ui/
│   │   │   ├── model/
│   │   │   ├── api/
│   │   │   └── lib/
│   │   ├── auth/                # Auth feature
│   │   │   ├── ui/
│   │   │   ├── model/
│   │   │   ├── api/
│   │   │   └── lib/
│   │   └── stores/              # Store selection feature
│   │       ├── ui/
│   │       ├── model/
│   │       └── api/
│   │
│   ├── entities/                # Business entities
│   │   ├── user/
│   │   │   ├── ui/
│   │   │   ├── model/
│   │   │   └── api/
│   │   ├── product/
│   │   │   ├── ui/
│   │   │   ├── model/
│   │   │   └── api/
│   │   ├── message/
│   │   │   ├── ui/
│   │   │   ├── model/
│   │   │   └── api/
│   │   └── store/
│   │       ├── ui/
│   │       ├── model/
│   │       └── api/
│   │
│   └── shared/                  # Shared resources
│       ├── ui/                  # UI Kit (Atomic Design)
│       │   ├── atoms/           # Basic elements
│       │   │   ├── Button/
│       │   │   ├── Input/
│       │   │   ├── Text/
│       │   │   ├── Icon/
│       │   │   └── Spinner/
│       │   ├── molecules/       # Composite elements
│       │   │   ├── FormField/
│       │   │   ├── SearchBar/
│       │   │   ├── MessageBubble/
│       │   │   └── ProductCard/
│       │   ├── organisms/       # Complex components
│       │   │   ├── MessageList/
│       │   │   ├── ProductList/
│       │   │   └── ChatInput/
│       │   └── templates/       # Page templates
│       │       ├── ChatLayout/
│       │       └── AuthLayout/
│       ├── api/                 # API clients
│       │   ├── client.ts
│       │   └── endpoints.ts
│       ├── lib/                 # Utilities
│       │   ├── hooks/
│       │   ├── utils/
│       │   └── constants/
│       └── config/              # Shared configs
│           ├── stores.ts
│           └── routes.ts
│
├── public/                      # Static assets
├── tests/                       # Test files
└── package.json
```

## 🏗️ Layer Description

### 1. **App Layer** (`src/app/`)
- Global providers (Theme, Auth, Redux)
- App-wide configurations
- Root-level styles

### 2. **Processes Layer** (`src/processes/`)
- Cross-cutting business flows
- Multi-step operations
- Complex user journeys

### 3. **Pages Layer** (`src/pages/`)
- Compositional layer
- Combines widgets and features
- Route-specific logic

### 4. **Widgets Layer** (`src/widgets/`)
- Large, self-contained UI sections
- Combines features and entities
- Examples: Header, Sidebar, ChatArea

### 5. **Features Layer** (`src/features/`)
- User interactions
- Business features
- Examples: send message, search products

### 6. **Entities Layer** (`src/entities/`)
- Business entities
- Domain objects
- Examples: User, Product, Message

### 7. **Shared Layer** (`src/shared/`)
- Reusable UI components (Atomic Design)
- API clients
- Utilities and helpers
- Common types

## 🧩 Component Architecture (Atomic Design)

### Atoms
```typescript
// src/shared/ui/atoms/Button/Button.tsx
interface ButtonProps {
  variant?: 'primary' | 'secondary' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
  children: React.ReactNode;
  onClick?: () => void;
  disabled?: boolean;
}

export const Button: FC<ButtonProps> = ({ 
  variant = 'primary',
  size = 'md',
  children,
  ...props 
}) => {
  return (
    <button 
      className={cn(
        'rounded-md font-medium transition-colors',
        variants[variant],
        sizes[size]
      )}
      {...props}
    >
      {children}
    </button>
  );
};
```

### Molecules
```typescript
// src/shared/ui/molecules/SearchBar/SearchBar.tsx
export const SearchBar: FC<SearchBarProps> = ({ 
  onSearch,
  placeholder 
}) => {
  return (
    <div className="relative">
      <Input 
        placeholder={placeholder}
        onChange={handleChange}
        icon={<SearchIcon />}
      />
      <Button 
        variant="ghost" 
        size="sm"
        onClick={handleSearch}
      >
        Search
      </Button>
    </div>
  );
};
```

### Organisms
```typescript
// src/shared/ui/organisms/MessageList/MessageList.tsx
export const MessageList: FC<MessageListProps> = ({ 
  messages 
}) => {
  return (
    <div className="space-y-4">
      {messages.map(message => (
        <MessageBubble 
          key={message.id}
          message={message}
          actions={<MessageActions message={message} />}
        />
      ))}
    </div>
  );
};
```

## 🔄 State Management

### Zustand for Features
```typescript
// src/features/chat/model/store.ts
export const useChatStore = create<ChatState>((set) => ({
  messages: [],
  isLoading: false,
  
  sendMessage: async (content: string) => {
    set({ isLoading: true });
    const message = await chatApi.send(content);
    set(state => ({ 
      messages: [...state.messages, message],
      isLoading: false 
    }));
  }
}));
```

### React Query for Server State
```typescript
// src/features/search/api/hooks.ts
export const useProductSearch = (query: string) => {
  return useQuery({
    queryKey: ['products', query],
    queryFn: () => searchApi.searchProducts(query),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};
```

## 🧪 Testing Strategy

### 1. **Unit Tests** (Atoms & Molecules)
```typescript
// src/shared/ui/atoms/Button/Button.test.tsx
describe('Button', () => {
  it('renders with correct variant', () => {
    render(<Button variant="primary">Click me</Button>);
    expect(screen.getByRole('button')).toHaveClass('bg-primary');
  });
});
```

### 2. **Integration Tests** (Features)
```typescript
// src/features/chat/tests/sendMessage.test.tsx
describe('Send Message Feature', () => {
  it('sends message and updates chat', async () => {
    renderWithProviders(<ChatFeature />);
    
    const input = screen.getByPlaceholderText('Type a message...');
    const button = screen.getByRole('button', { name: 'Send' });
    
    await userEvent.type(input, 'Hello ShopGPT');
    await userEvent.click(button);
    
    expect(await screen.findByText('Hello ShopGPT')).toBeInTheDocument();
  });
});
```

### 3. **E2E Tests** (User Flows)
```typescript
// tests/e2e/shopping-flow.test.ts
test('complete shopping flow', async ({ page }) => {
  await page.goto('/');
  await page.fill('[placeholder="Search for any product..."]', 'laptop');
  await page.click('button[type="submit"]');
  
  await expect(page.locator('.product-card')).toHaveCount(10);
  await page.click('.product-card:first-child');
  
  await expect(page).toHaveURL(/\/products\/\d+/);
});
```

## 📋 Best Practices

### 1. **Import Rules**
- Features cannot import from other features
- Shared cannot import from any layer above
- Use barrel exports for clean imports

### 2. **Component Guidelines**
- Keep components pure and testable
- Use composition over inheritance
- Implement proper error boundaries

### 3. **Performance Optimization**
- Lazy load features and pages
- Use React.memo for expensive components
- Implement virtual scrolling for lists

### 4. **Type Safety**
- Use TypeScript strictly
- Define clear interfaces
- Avoid `any` types

## 🔗 API Integration

### API Client Setup
```typescript
// src/shared/api/client.ts
export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  timeout: 10000,
});

apiClient.interceptors.request.use((config) => {
  const token = authStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

### Feature API Layer
```typescript
// src/features/search/api/searchApi.ts
export const searchApi = {
  searchProducts: async (query: string, store?: string) => {
    const { data } = await apiClient.get('/search', {
      params: { q: query, store }
    });
    return data;
  },
  
  getSuggestions: async (query: string) => {
    const { data } = await apiClient.get('/suggestions', {
      params: { q: query }
    });
    return data;
  }
};
```

## 🚀 Development Workflow

1. **Component Development**
   - Start with atoms in Storybook
   - Compose into molecules and organisms
   - Test in isolation

2. **Feature Development**
   - Define feature model (state, types)
   - Implement UI components
   - Add API integration
   - Write tests

3. **Integration**
   - Combine features in widgets
   - Compose widgets in pages
   - Test complete user flows

## 📚 Resources

- [Feature-Sliced Design](https://feature-sliced.design/)
- [Atomic Design by Brad Frost](https://atomicdesign.bradfrost.com/)
- [React Testing Library](https://testing-library.com/react)
- [Zustand Documentation](https://zustand-demo.pmnd.rs/)