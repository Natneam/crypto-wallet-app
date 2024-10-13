import React, { useState, useEffect } from "react";
import "./Dashboard.css";

const API_URL = process.env.REACT_APP_API_URL;

const getAuthToken = () => {
  return localStorage.getItem("token");
};

export default function Dashboard({ logout }) {
  const [wallets, setWallets] = useState([]);
  const [newWalletName, setNewWalletName] = useState("");

  const fetchWallets = async () => {
    try {
      const response = await fetch(`${API_URL}/api/wallets`, {
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
        },
      });
      if (response.status === 401) {
        logout();
        return;
      }

      if (!response.ok) {
        alert("Failed to fetch wallets");
        throw new Error("Failed to fetch wallets");
      }
      const data = await response.json();
      setWallets(data);
    } catch (error) {
      console.error("Error fetching wallets:", error);
    }
  };

  useEffect(() => {
    fetchWallets();
  }, []);

  const handleCreateWallet = async (e) => {
    e.preventDefault();
    try {
      const response = await fetch(`${API_URL}/api/wallet`, {
        method: "POST",
        body: JSON.stringify({ name: newWalletName }),
        headers: {
          Authorization: `Bearer ${getAuthToken()}`,
          "Content-Type": "application/json",
        },
      });

      if (response.status === 401) {
        logout();
        return;
      }

      if (!response.ok) {
        alert("Failed to create wallet");
        throw new Error("Failed to create wallet");
      }
      setNewWalletName("");
      fetchWallets();
    } catch (error) {
      console.error("Error creating wallet:", error);
    }
  };

  return (
    <div className="dashboard-container">
      <h2 className="dashboard-title">Wallet Dashboard</h2>

      <div className="create-wallet-section">
        <h3>Create New Wallet</h3>
        <form onSubmit={handleCreateWallet} className="create-wallet-form">
          <input
            type="text"
            required={true}
            value={newWalletName}
            onChange={(e) => setNewWalletName(e.target.value)}
            placeholder="Enter wallet name"
            className="wallet-input"
          />
          <button type="submit" className="button">
            Create Wallet
          </button>
        </form>
      </div>

      <div className="wallet-list-section">
        <div className="wallet-list-header">
          <h3>Wallet List</h3>
          <button onClick={fetchWallets} className="button">
            Refresh
          </button>
        </div>
        <ul className="wallet-list">
          {wallets?.map((wallet) => (
            <li key={wallet.public_key} className="wallet-item">
              <strong>{wallet.name}</strong>
              <br />
              Public Key: {wallet.public_key}
              <br />
              Balance: {wallet.balance}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
