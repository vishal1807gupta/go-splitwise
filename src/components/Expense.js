import React, { useState } from "react";
import DropdownMenu from "./DropDownMenu";

const Expense = ({ users, groupId, onExpenseAdded, onCancel }) => {
  const [checkedUsers, setCheckedUsers] = useState([]);
  const [expenseType, setExpenseType] = useState("");
  const [payerId, setPayerId] = useState("");
  const [amount, setAmount] = useState("");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [toast, setToast] = useState({ show: false, message: "", type: "" });

  if (!users) return null;

  const findUserName = (userId) => {
    const user = users.find(u => u.id === userId);
    return user ? user.name : `User ${userId}`;
  };

  // Show toast notification instead of alert
  const showToast = (message, type = "error") => {
    setToast({ show: true, message, type });
    
    // Automatically hide toast after 5 seconds
    setTimeout(() => {
      setToast({ show: false, message: "", type: "" });
    }, 5000);
  };

  const validateExpense = (description, amount, payerId, expenseType, checkedUsers) => {
      if (!description || description.trim().length < 3) {
        return "Description must be at least 3 characters long.";
      }
      const amountNum = parseInt(amount);
      if (isNaN(amountNum) || amountNum <= 0) {
        return "Amount must be a positive number.";
      }
      if (!payerId) {
        return "Please select a payer for this expense.";
      }
      if (!expenseType) {
        return "Please select an expense type.";
      }
      if (checkedUsers.length === 0) {
        return "Please select at least one user to share the expense with.";
    }
    
    return null;
  };

  const validateShares = (expenseType, totalAmount, checkedUsers) => {
    if(expenseType==="EQUAL")return;
    let sum = 0;
    let error = 0;
    checkedUsers.forEach((userId) => {
      const shareInput = document.getElementById(`amount-${userId}`);
      const shareAmount = parseInt(shareInput.value);
      
      if (isNaN(shareAmount) || shareAmount <= 0) {
        error=userId;
      }
      
      sum += shareAmount;
    });

    if(error)return `Invalid share amount for user ${findUserName(error)}. Must be a positive integer.`;
    
    if (expenseType === "EXACT") {
      const totalAmountNum = parseInt(totalAmount);
      if (sum !== totalAmountNum) {
        return `The sum of share amounts ${sum.toFixed(2)} must equal the total amount ${totalAmountNum.toFixed(2)}.`;
      }
    } 
    else if (expenseType === "PERCENTAGE") {
      if (sum!==100) {
        return `The sum of percentages (${sum.toFixed(2)}) must equal 100%.`;
      }
    }
    
    return null;
  };

  const handleSubmit = async () => {
    const error = validateExpense(description, amount, payerId, expenseType, checkedUsers);
    if (error) {
      showToast(error);
      return;
    }
    
    const sharesError = validateShares(expenseType, amount, checkedUsers);
    if (sharesError) {
      showToast(sharesError);
      return;
    }

    setIsLoading(true);

    const bodyData = {
      amount: parseInt(amount),
      payer_id: payerId,
      description: description,
      expense_type: expenseType,
      user_shares: checkedUsers.map((userId) => ({
        user_id: userId,
        share_amount: Number(document.getElementById(`amount-${userId}`).value),
      })),
    };

    try {
      const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/addExpense/${groupId}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(bodyData),
      });
      
      if (!response.ok) {
        throw new Error('Failed to add expense');
      }
      
      const data = await response.json();
      console.log(data);
      
      // Call the callback if provided
      if (onExpenseAdded) {
        onExpenseAdded();
      } else {
        showToast("Expense added successfully!", "success");
        
        // Reset form if not being closed
        setCheckedUsers([]);
        setAmount("");
        setDescription("");
      }
      
    } catch (error) {
      showToast(error.message || "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  const toggle = (userId) => {
    setCheckedUsers((prevCheckedUsers) => {
      if (prevCheckedUsers.includes(userId)) {
        return prevCheckedUsers.filter((id) => id !== userId);
      } else {
        return [...prevCheckedUsers, userId];
      }
    });
  };

  return (
    <div className="space-y-3 relative">
      {/* Toast Notification */}
      {toast.show && (
        <div 
          className={`fixed top-4 right-4 z-50 p-4 rounded-md shadow-lg max-w-md animate-fade-in ${
            toast.type === "success" ? "bg-green-50 border-l-4 border-green-500" : "bg-red-50 border-l-4 border-red-500"
          }`}
        >
          <div className="flex items-start">
            <div className="flex-shrink-0">
              {toast.type === "success" ? (
                <svg className="h-5 w-5 text-green-600" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              ) : (
                <svg className="h-5 w-5 text-red-600" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              )}
            </div>
            <div className="ml-3">
              <p className={`text-sm font-medium ${toast.type === "success" ? "text-green-800" : "text-red-800"}`}>
                {toast.message}
              </p>
            </div>
            <div className="ml-auto pl-3">
              <div className="-mx-1.5 -my-1.5">
                <button
                  onClick={() => setToast({ show: false, message: "", type: "" })}
                  className={`inline-flex rounded-md p-1.5 ${
                    toast.type === "success" 
                      ? "bg-green-50 text-green-500 hover:bg-green-100" 
                      : "bg-red-50 text-red-500 hover:bg-red-100"
                  }`}
                >
                  <svg className="h-4 w-4" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Two columns for top form elements */}
      <div className="grid grid-cols-2 gap-3">
        {/* Description - Full width */}
        <div className="col-span-2">
          <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
            Description
          </label>
          <input 
            type="text" 
            id="description" 
            placeholder="What was this expense for?" 
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none"
          />
        </div>
        
        {/* Amount - First column */}
        <div>
          <label htmlFor="expense-amount" className="block text-sm font-medium text-gray-700 mb-1">
            Total Amount
          </label>
          <input 
            type="number" 
            id="expense-amount" 
            placeholder="0.00" 
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none"
          />
        </div>
        
        {/* Who paid? - Second column */}
        <div>
          <label htmlFor="payer" className="block text-sm font-medium text-gray-700 mb-1">
            Who Paid?
          </label>
          <DropdownMenu 
            options={users} 
            dropdownId="payer" 
            onSelect={(value) => setPayerId(value)}
          />
        </div>
      </div>
      
      {/* Split Type */}
      <div>
        <label htmlFor="expenseType" className="block text-sm font-medium text-gray-700 mb-1">
          Split Type
        </label>
        <DropdownMenu 
          options={[
            { label: "Split Equally", value: "EQUAL" },
            { label: "Exact Amounts", value: "EXACT" },
            { label: "Percentages", value: "PERCENTAGE" }
          ]} 
          dropdownId="expenseType"
          onSelect={(value) => setExpenseType(value)}
        />
      </div>
      
      {/* Users to split with */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">
          Split With <span className="text-xs text-gray-500">({users.length} members)</span>
        </label>
        <div className="border border-gray-200 rounded-md">
          <div className="max-h-32 overflow-y-auto p-2">
            {users.map((user) => (
              <div key={user.id} className="flex items-center justify-between p-2 border border-gray-100 rounded hover:bg-gray-50 mb-2 last:mb-0">
                <div className="flex items-center flex-grow">
                  <input 
                    type="checkbox" 
                    id={`check-${user.id}`}
                    checked={checkedUsers.includes(user.id)}
                    onChange={() => toggle(user.id)} 
                    className="h-4 w-4 text-indigo-600 rounded mr-3"
                  />
                  <div className="flex-grow">
                    <div className="text-sm font-medium">{user.name}</div>
                    <div className="text-xs text-gray-500">{user.email}</div>
                  </div>
                </div>
                
                <input 
                  type="number" 
                  id={`amount-${user.id}`}
                  placeholder="0.00" 
                  disabled={!checkedUsers.includes(user.id)}
                  className="w-24 px-2 py-1 text-sm border border-gray-300 rounded shadow-sm"
                />
              </div>
            ))}
          </div>
        </div>
      </div>
      
      {/* Action Buttons */}
      <div className="flex justify-end space-x-3 pt-2">
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
          >
            Cancel
          </button>
        )}
        <button 
          onClick={handleSubmit}
          disabled={isLoading}
          className={`px-4 py-2 bg-indigo-600 text-white font-medium rounded-md hover:bg-indigo-700 focus:outline-none ${onCancel ? '' : 'w-full'}`}
        >
          {isLoading ? (
            <>
              <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white inline" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Processing...
            </>
          ) : (
            'Add Expense'
          )}
        </button>
      </div>
    </div>
  );
};

export default Expense;