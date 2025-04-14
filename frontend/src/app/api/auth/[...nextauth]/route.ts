import NextAuth from "next-auth";
import GoogleProvider from "next-auth/providers/google";
import axios from "axios";

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Function to exchange Google token for backend JWT
const exchangeGoogleToken = async (googleToken: string): Promise<string | null> => {
  try {
    console.log("Exchanging Google token for backend JWT");
    const response = await axios.post(`${API_URL}/auth/token`, { token: googleToken }, {
      headers: {
        'Content-Type': 'application/json',
      }
    });
    
    if (response.data && response.data.token) {
      console.log("Successfully received backend JWT token");
      return response.data.token;
    }
    console.error("Backend response missing token:", response.data);
    return null;
  } catch (error: any) {
    console.error("Failed to exchange Google token:", error);
    if (error.response) {
      console.error("Error response data:", error.response.data);
      console.error("Error response status:", error.response.status);
    }
    return null;
  }
};

const handler = NextAuth({
  providers: [
    GoogleProvider({
      clientId: process.env.GOOGLE_CLIENT_ID!,
      clientSecret: process.env.GOOGLE_CLIENT_SECRET!,
    }),
  ],
  callbacks: {
    async jwt({ token, account }) {
      // Persist the OAuth access_token to the token right after signin
      if (account) {
        token.accessToken = account.access_token;
        
        // Try to exchange Google token for backend JWT
        if (account.access_token) {
          try {
            const backendToken = await exchangeGoogleToken(account.access_token);
            if (backendToken) {
              token.backendToken = backendToken;
              console.log("Added backend JWT to session token");
            }
          } catch (error) {
            console.error("Error exchanging token in jwt callback:", error);
          }
        }
      }
      return token;
    },
    async session({ session, token }) {
      // Send properties to the client
      session.accessToken = token.accessToken as string;
      if (token.backendToken) {
        session.backendToken = token.backendToken as string;
      }
      return session;
    },
  },
  pages: {
    signIn: "/auth/signin",
  },
});

export { handler as GET, handler as POST }; 