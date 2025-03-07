function toggleReplyForm(commentId) {
    const replyForm = document.getElementById(`reply-form-${commentId}`);
    if (replyForm) {
        const isVisible = replyForm.style.display !== 'none';
        replyForm.style.display = isVisible ? 'none' : 'block';
        
        // 如果打开了当前回复框，关闭其他所有回复框
        if (!isVisible) {
            const allReplyForms = document.querySelectorAll('.reply-form');
            allReplyForms.forEach(form => {
                if (form.id !== `reply-form-${commentId}`) {
                    form.style.display = 'none';
                }
            });
        }
    }
}