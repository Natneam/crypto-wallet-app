import React, { useState } from "react";
import "./SendFund.css";

const SendFundsPopup = ({ isOpen, onClose, onSend, walletName }) => {
  const [recipientAddress, setRecipientAddress] = useState("");
  const [amount, setAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [transactionData, setTransactionData] = useState(null);

  const handleSend = async () => {
    setLoading(true);
    const result = await onSend(recipientAddress, amount);
    if (!result) {
      console.log("Failed to send funds");
      setLoading(false);
      return;
    }
    setTransactionData(result);
    setLoading(false);
  };

  const handleClose = () => {
    setAmount("");
    setRecipientAddress("");
    setTransactionData(null);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="popup-overlay">
      <div className="popup-content">
        {!transactionData ? (
          <>
            <h2 className="popup-title">Send Funds</h2>
            <h3 className="popup-subtitle">From: {walletName}</h3>
            <input
              type="text"
              required={true}
              placeholder="Recipient Wallet Address"
              value={recipientAddress}
              onChange={(e) => setRecipientAddress(e.target.value)}
              className="popup-input"
            />
            <input
              type="number"
              placeholder="Amount in Wei"
              required={true}
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              className="popup-input"
            />
            <div className="popup-buttons">
              <button
                onClick={handleSend}
                className="popup-button popup-button-primary"
                disabled={loading}
              >
                {loading ? "Sending..." : "Send"}
              </button>
              <button
                onClick={handleClose}
                className="popup-button popup-button-secondary"
              >
                Cancel
              </button>
            </div>
          </>
        ) : (
          <>
            <div className="popup-header">
              <h2 className="popup-title">Transaction Successful</h2>
              <button onClick={handleClose} className="popup-close-button">&times;</button>
            </div>
            <div className="popup-success-message">
              <span className="popup-success-icon">&#10004;</span>
              <p>Transaction completed successfully!</p>
            </div>
            <div className="popup-transaction-details">
              <p><strong>Transaction Hash:</strong> {transactionData.transactionHash}</p>
              <div className="popup-transaction-addresses">
                <p><strong>From:</strong> {transactionData.from}</p>
                <span className="popup-arrow">&rarr;</span>
                <p><strong>To:</strong> {transactionData.to}</p>
              </div>
              <p><strong>Value:</strong> {transactionData.value} Wei</p>
              <p><strong>Gas Price:</strong> {transactionData.gasPrice} Wei</p>
              <p><strong>Gas Used:</strong> {transactionData.gasUsed}</p>
              <p><strong>Block Number:</strong> {transactionData.blockNumber}</p>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default SendFundsPopup;