import React, { useState } from "react";
import SendFundsPopup from "../SendFunds/SendFund";
import "./Wallet.css";

const API_URL = process.env.REACT_APP_API_URL;

const getAuthToken = () => {
  return localStorage.getItem("token");
};

async function sendFund(fromWallet, toWallet, amount, logout) {
  const response = await fetch(`${API_URL}/api/sign-transaction`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
    body: JSON.stringify({
      fromAddress: fromWallet,
      toAddress: toWallet,
      value: amount,
    }),
  });

  if(response.status === 401) {
    logout();
    return;
  }

  if (!response.ok) {
    const errorData = await response.json();
    alert(errorData.message || "Error sending fund");
    return;
  }
  return await response.json();
}

const WalletItem = ({ wallet, logout }) => {
  const [isPopupOpen, setIsPopupOpen] = useState(false);

  const handleSendFunds = async (recipientAddress, amount) => {
    const result = await sendFund(wallet.public_key, recipientAddress, amount, logout);
    if (!result) {
      console.log("Failed to send funds");
      return;
    }

    return result;
  };

  return (
    <li className="wallet-item">
      <div className="wallet-content">
        <div className="wallet-details">
          <strong>{wallet.name}</strong>
          <br />
          Public Key: {wallet.public_key}
          <br />
          Balance: {wallet.balance}
        </div>
        <div className="wallet-actions">
          <button
            onClick={() => setIsPopupOpen(true)}
            className="send-funds-button"
          >
            Send Funds
          </button>
        </div>
      </div>
      <SendFundsPopup
        isOpen={isPopupOpen}
        onClose={() => setIsPopupOpen(false)}
        onSend={handleSendFunds}
        walletName={wallet.name}
      />
    </li>
  );
};

export default WalletItem;
