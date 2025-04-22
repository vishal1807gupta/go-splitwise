// src/services/cognitoService.js
import { CognitoIdentityProviderClient, ForgotPasswordCommand, ConfirmForgotPasswordCommand } from "@aws-sdk/client-cognito-identity-provider";

const REGION = "ap-southeast-2";
const CLIENT_ID = process.env.REACT_APP_COGNITO_CLIENT_ID;

const client = new CognitoIdentityProviderClient({ region: REGION });

export const forgotPassword = async (email) => {
  const command = new ForgotPasswordCommand({
    ClientId: CLIENT_ID,
    Username: email,
  });

  try {
    const response = await client.send(command);
    return response;
  } catch (error) {
    console.error("Error in forgotPassword:", error);
    throw error;
  }
};

export const confirmForgotPassword = async (email, code, newPassword) => {
  const command = new ConfirmForgotPasswordCommand({
    ClientId: CLIENT_ID,
    Username: email,
    ConfirmationCode: code,
    Password: newPassword,
  });

  try {
    const response = await client.send(command);
    return response;
  } catch (error) {
    console.error("Error in confirmForgotPassword:", error);
    throw error;
  }
};

export const updatePasswordInDatabase = async (email, newPassword) => {
    try {
      const response = await fetch('http://localhost:4000/api/update-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, newPassword }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to update password in database');
      }
      
      return await response.json();
    } catch (error) {
      console.error("Error updating password in database:", error);
      throw error;
    }
  };