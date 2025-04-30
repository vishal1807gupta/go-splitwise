import React, { createContext, useState, useEffect, useContext } from "react";
import axios from 'axios';


// Create the authentication context
const AuthContext = createContext(null);

// Custom hook to use the auth context
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

// Auth context provider component
export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [loading, setLoading] = useState(true);

  // Check if user is logged in on mount
  useEffect(() => {
    const checkLoggedIn = async () => {
      try {
        const response = await fetch("http://localhost:4000/api/me", {
          method: "GET",
          credentials: "include", // Important for cookies
        });

        if (response.ok) {
          const data = await response.json();
          setCurrentUser(data);
        } else {
          setCurrentUser(null);
        }
      } catch (error) {
        console.error("Error checking login status:", error);
        setCurrentUser(null);
      } finally {
        setLoading(false);
      }
    };

    checkLoggedIn();
  }, []);

  // Login function
  const login = async (email, password, rememberMe) => {
    const response = await fetch("http://localhost:4000/api/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password, rememberMe }),
      credentials: "include",
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || "Login failed");
    }
    console.log(data);
    setCurrentUser(data);
    return data;
  };

  // Register function
  const register = async (name, email, password) => {
    const response = await fetch("http://localhost:4000/api/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, email, password }),
      credentials: "include",
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || "Registration failed");
    }
    console.log(data);
    setCurrentUser(data);
    return data;
  };

  // Logout function
  const logout = async () => {
    try {
      await fetch("http://localhost:4000/api/logout", {
        method: "POST",
        credentials: "include",
      });// needs to be implemented like blacklisting jwt token, remove session information from server to avoid further api requests
      setCurrentUser(null);
    } catch (error) {
      console.error("Logout failed:", error);
    }
  };

  const requestPasswordReset = async (email) => {
    try {
      const response = await axios.post('http://localhost:4000/api/auth/request-password-reset', { email });
      console.log(response);
      return response.data;
    } catch (error) {
      throw new Error(error.response?.data || 'Failed to request password reset');
    }
  };

  const resetPasswordComplete = async (email, code, newPassword) => {
    try {
      const response = await axios.post('http://localhost:4000/api/auth/reset-password-complete', {
        email,
        code,
        newPassword
      });
      console.log(response);
      return response.data;
    } catch (error) {
      throw new Error(error.response?.data || 'Failed to reset password');
    }
  };


  const value = {
    currentUser,
    loading,
    login,
    register,
    logout,
    setCurrentUser,
    resetPasswordComplete,
    requestPasswordReset,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};