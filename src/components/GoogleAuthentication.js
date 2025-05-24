import React, { useState } from "react";
import { GoogleLogin } from '@react-oauth/google';
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

const GoogleAuthentication = () => {
    const navigate = useNavigate();
    const { setCurrentUser } = useAuth();
    const [isLoading, setIsLoading] = useState(false);

    const handleGoogleSuccess = async (credentialResponse) => {
        try {
            setIsLoading(true);
            const token = credentialResponse.credential;
            console.log("Got token:", token);
            
            const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/auth/google`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ token }),
                credentials: 'include'  // decides whether to store cookies or reject them
            });
            
            const data = await response.json();
            console.log("Login success:", data);

            if (data) {
                setCurrentUser(data);
                navigate('/groups'); // Redirect to groups page
            }
            
            // Handle successful login (e.g., redirect or update state)
        } catch (error) {
            console.error("Login error:", error);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="relative">
            <div className="flex justify-center">
                <GoogleLogin
                    onSuccess={handleGoogleSuccess}
                    onError={(error) => {
                        console.error('Login Failed:', error);
                    }}
                    theme="filled_blue"
                    disabled={isLoading}
                />
            </div>
            
            {/* Loading Overlay */}
            {isLoading && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                    <div className="animate-spin rounded-full h-12 w-12 border-4 border-white border-t-transparent"></div>
                </div>
            )}
        </div>
    );
};

export default GoogleAuthentication;