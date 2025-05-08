import React, { createContext, useState, useEffect, useContext } from "react";
import axios from 'axios';

const AuthContext = createContext(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const validateName = (name) => {
    if (!name || name.trim() === '') {
      return "Name is required.";
    }
    if (name.trim().length < 3) {
      return "Name must be at least 3 characters long.";
    }
    return null;
  };
  
  const validateEmail = (email) => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!email || email.trim() === '') {
      return "Email is required.";
    }
    if (!emailRegex.test(email)) {
      return "Please enter a valid email address.";
    }
    return null;
  };

  const validatePassword = (password) => {
    if (!password) {
      return "Password is required.";
    }
    if (password.length < 6) {
      return "Password must be at least 6 characters long.";
    }
    return null;
  };

  useEffect(() => {
    const checkLoggedIn = async () => {
      try {
        const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/me`, {
          method: "GET",
          credentials: "include", // Important for cookies
        });

        console.log(response);

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

  const login = async (email, password, rememberMe) => {
    try {
      const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, password, rememberMe }),
        credentials: "include",
      });

      if (!response.ok) {
        if (response.status === 401) {
          throw new Error("Incorrect email or password. Please try again.");
        } else {
          throw new Error("Unable to log in. Please try again later.");
        }
      }

      const data = await response.json();
      setCurrentUser(data);
      return data;
    } catch (error) {
      throw error;
    }
  };

  const register = async (name, email, password) => {
    const nameError = validateName(name);
    if (nameError) throw new Error(nameError);
    
    const emailError = validateEmail(email);
    if (emailError) throw new Error(emailError);
    
    const passwordError = validatePassword(password);
    if (passwordError) throw new Error(passwordError);
    try {
      const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/register`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name, email, password }),
        credentials: "include",
      });

      if (!response.ok) {
        if (response.status === 409) {
          throw new Error("This email is already registered. Please use a different email or try logging in.");
        } else if (response.status === 400) {
          throw new Error("Please check your information and try again.");
        } else {
          throw new Error("Registration failed. Please try again later.");
        }
      }
      
      const data = await response.json();
      setCurrentUser(data);
      return data;
    } catch (error) {
      throw error;
    }
  };

  const logout = async () => {
    try {
      const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/logout`, {
        method: "POST",
        credentials: "include",
      });
      console.log(response);
      setCurrentUser(null);
    } catch (error) {
      console.error("Logout failed:", error);
      throw new Error("Logout failed. Please refresh the page and try again.");
    }
  };

  const requestPasswordReset = async (email) => {
    const emailError = validateEmail(email);
    if (emailError) throw new Error(emailError);
    try {
      const response = await axios.post(`${process.env.REACT_APP_BACKEND_URL}/api/auth/request-password-reset`, { email });
      return response.data;
    } catch (error) {
      if (error.response?.status === 404) {
        throw new Error("No account found with this email address.");
      } else if (error.response?.status === 429) {
        throw new Error("Too many reset requests. Please try again later.");
      } else {
        throw new Error(error.response?.data?.message || "Could not send reset email. Please try again later.");
      }
    }
  };

  const resetPasswordComplete = async (email, code, newPassword) => {

    const emailError = validateEmail(email);
    if (emailError) throw new Error(emailError);
    
    const passwordError = validatePassword(newPassword);
    if (passwordError) throw new Error(passwordError);
    
    if (!code || code.trim().length<6) {
      throw new Error("Empty or Incorrect Verification code.");
    }

    try {
      const response = await axios.post(`${process.env.REACT_APP_BACKEND_URL}/api/auth/reset-password-complete`, {
        email,
        code,
        newPassword
      });
      return response.data;
    } catch (error) {
      if (error.response?.status === 400) {
        throw new Error(error.response?.data);
      } else {
        throw new Error(error.response?.data?.message || "Could not reset password. Please try again.");
      }
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