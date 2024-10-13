import React from "react";
import "./App.css";
import { BrowserRouter, Route, Routes } from "react-router-dom";
import Login from './components/Login/Login';
import Dashboard from "./components/Dashboard/Dashboard";
import NotFound from "./components/NotFound/NotFound";
import Signup from "./components/Signup/Signup";
import Layout from "./components/Layout/Layout";
import useToken from './useToken';

const logout = (setToken) => {
  setToken(null);
  localStorage.removeItem("token");
};

function App() {
  const { token, setToken } = useToken();

  return (
    <div className="wrapper">
      <BrowserRouter>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<Login setToken={setToken} />} />
          <Route path="/signup" element={<Signup />} />

          {/* Protected routes */}
          <Route 
            path="/" 
            element={
              token ? (
                <Layout logout={() => logout(setToken)}>
                  <Dashboard logout={() => logout(setToken)} />
                </Layout>
              ) : (
                <Login setToken={setToken} />
              )
            } 
          />

          {/* 404 route */}
          <Route path="*" element={<NotFound />} />
        </Routes>
      </BrowserRouter>
    </div>
  );
}

export default App;