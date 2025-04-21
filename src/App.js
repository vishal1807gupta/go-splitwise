import React from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import Groups from "./components/Groups";
import Group from "./components/Group";
import { AuthProvider } from "./auth/AuthContext";
import Login from "./components/Login";
import Register from "./components/Register";
import ProtectedRoute from "./components/ProtectedRoute";
import Navbar from "./components/Navbar";

const App = () => {
  return (
    <>
      <Router>
        <AuthProvider>
          <div className="flex flex-col min-h-screen">
            <Navbar />
            <div className="flex-grow mt-16">
              <Routes>
                {/* Public routes */}
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />
                
                {/* Protected routes */}
                <Route element={<ProtectedRoute />}>
                  <Route path="/groups" element={<Groups />} />
                  <Route path="/groups/:groupId/:groupName" element={<Group />} />
                  {/* Redirect to groups page if root path is accessed */}
                  <Route path="/" element={<Navigate replace to="/groups" />} />
                </Route>
                
                {/* Fallback route for unknown paths */}
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </div>
          </div>
        </AuthProvider>
      </Router>
    </>
  );
};

export default App;