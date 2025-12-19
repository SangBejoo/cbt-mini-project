'use client';

import { ChakraProvider, extendTheme } from '@chakra-ui/react';
import { CacheProvider } from '@emotion/react';
import { clientSideEmotionCache } from '../utils/emotion-cache';

const theme = extendTheme({});

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <CacheProvider value={clientSideEmotionCache}>
      <ChakraProvider theme={theme}>{children}</ChakraProvider>
    </CacheProvider>
  );
}