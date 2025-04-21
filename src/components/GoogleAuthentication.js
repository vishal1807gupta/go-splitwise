import React from "react";
import { GoogleLogin } from '@react-oauth/google';
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

const GoogleAuthentication = () => {
    const navigate = useNavigate();
    const { setCurrentUser } = useAuth();

    const handleGoogleSuccess = async (credentialResponse) => {
        try {
            const token = credentialResponse.credential;
            console.log("Got token:", token);
            
            const response = await fetch('http://localhost:4000/api/auth/google', {
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
        }
    };

    return (
        <div className="flex justify-center">
            <GoogleLogin
                onSuccess={handleGoogleSuccess}
                onError={(error) => {
                    console.error('Login Failed:', error);
                }}
                theme="filled_blue"
            />
        </div>
    );
};

export default GoogleAuthentication;